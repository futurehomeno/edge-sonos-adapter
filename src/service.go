package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/sonos/model"
	"github.com/thingsplex/sonos/router"
	"github.com/thingsplex/sonos/sonos-api"
)

func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := model.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	states := model.NewStates(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}
	err = states.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load state file.")
	}
	client := sonos.Client{}

	edgeapp.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting sonos----------------")
	log.Info("Work directory : ", configs.WorkDir)
	appLifecycle.PublishEvent(model.EventConfiguring, "main", nil)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()

	fimpRouter := router.NewFromFimpRouter(mqtt, appLifecycle, configs, states)
	fimpRouter.Start()

	appLifecycle.SetConnectionState(model.ConnStateDisconnected)
	appLifecycle.SetAppState(model.AppStateRunning, nil)
	if states.IsConfigured() && err == nil {
		appLifecycle.SetConfigState(model.ConfigStateConfigured)
		appLifecycle.SetConnectionState(model.ConnStateConnected)
	} else {
		appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
		appLifecycle.SetConnectionState(model.ConnStateDisconnected)
	}

	if configs.IsAuthenticated() && err == nil {
		appLifecycle.SetAuthState(model.AuthStateAuthenticated)
	} else {
		appLifecycle.SetAuthState(model.AuthStateNotAuthenticated)
	}

	for {
		appLifecycle.WaitForState("main", model.AppStateRunning)
		ticker := time.NewTicker(time.Duration(15) * time.Second)
		counter := 4
		for ; true; <-ticker.C {
			if configs.AccessToken != "" && configs.AccessToken != "access_token" {
				// ADD LOGIC TO HANDLE REFRESH TOKEN
				// every 24 hours at least
				// if millis is more than 12 hours after last authorization, make new
				millis := time.Now().UnixNano() / 1000000
				refreshMillis := configs.LastAuthMillis + 43200000
				if millis > refreshMillis {
					newAccessToken, err := client.RefreshAccessToken(configs.RefreshToken, configs.MqttServerURI, configs.Env)
					if err != nil {
						log.Error(err)
					}
					configs.AccessToken = newAccessToken
					configs.LastAuthMillis = millis
					configs.SaveToFile()
				}
			}

			counter++
			for i := 0; i < len(states.Groups); i++ {
				metadata, err := client.GetMetadata(configs.AccessToken, states.Groups[i].GroupId)
				if err == nil {
					states.Container = metadata.Container
					if states.Container.Service.Name == "Sonos Radio" {
						states.StreamInfo = metadata.StreamInfo
						states.Container.ImageURL = "https://static.vecteezy.com/system/resources/previews/000/581/923/non_2x/radio-icon-vector-illustration.jpg"
					} else if states.Container.Service.Name == "Spotify" || states.Container.Service.Name == "Apple Music" {
						states.CurrentItem = metadata.CurrentItem
						states.NextItem = metadata.NextItem
						states.StreamInfo = ""
					}

					report := map[string]interface{}{
						"album":       states.CurrentItem.Track.Album.Name,
						"track":       states.CurrentItem.Track.Name,
						"artist":      states.CurrentItem.Track.Artist,
						"image_url":   states.Container.ImageURL,
						"stream_info": states.StreamInfo,
					}
					adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: states.Groups[i].FimpId}
					msg := fimpgo.NewMessage("evt.metadata.report", "media_player", fimpgo.VTypeStrMap, report, nil, nil, nil)
					mqtt.Publish(adr, msg)
					if counter >= 4 {
						pbState, err := client.GetPlaybackStatus(states.Groups[i].GroupId, configs.AccessToken)
						if err != nil {
							break
						}
						volume, err := client.GetVolume(states.Groups[i].GroupId, configs.AccessToken)
						if err != nil {
							break
						}

						states.PlaybackState = pbState.PlaybackState
						states.PlayModes = pbState.PlayModes
						states.Volume = volume.Volume
						states.Muted = volume.Muted
						states.Fixed = volume.Fixed

						msg = fimpgo.NewMessage("evt.playback.report", "media_player", fimpgo.VTypeString, states.PlaybackState, nil, nil, nil)
						mqtt.Publish(adr, msg)
						msg = fimpgo.NewMessage("evt.mode.report", "media_player", fimpgo.VTypeStrMap, states.PlayModes, nil, nil, nil)
						mqtt.Publish(adr, msg)
						msg = fimpgo.NewMessage("evt.volume.report", "media_player", fimpgo.VTypeInt, states.Volume, nil, nil, nil)
						mqtt.Publish(adr, msg)
						msg = fimpgo.NewMessage("evt.mute.report", "media_player", fimpgo.VTypeBool, states.Muted, nil, nil, nil)
						mqtt.Publish(adr, msg)
						counter = 0
					}
					log.Debug(metadata.Container.ImageURL)
					log.Debug(states.Container.ImageURL)
				} else {
					log.Error("This is the one in service.go", err)
				}
			}
			log.Debug("ticker")
			states.SaveToFile()
		}
		appLifecycle.WaitForState(model.AppStateNotConfigured, "main")
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
