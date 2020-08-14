package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	"reflect"
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
	if funk.IsEmpty(workDir) {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}

	appLifecycle := model.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	states := model.NewStates(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't load config file."))
	}
	err = states.LoadFromFile()
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't load state file."))
	}
	client := sonos.NewClient(configs.Env)

	edgeapp.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting sonos----------------")
	log.Info("Work directory : ", configs.WorkDir)
	appLifecycle.PublishEvent(model.EventConfiguring, "main", nil)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	defer mqtt.Stop()

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
		var oldReport map[string]interface{}
		var oldPbStateValue string
		var oldPlayModes struct {
			Repeat    bool `json:"repeat"`
			RepeatOne bool `json:"repeatOne"`
			Shuffle   bool `json:"shuffle"`
			Crossfade bool `json:"crossfade"`
		}
		var oldVolume int
		var oldMuted bool

		for range ticker.C {
			if configs.AccessToken != "" && configs.AccessToken != "access_token" {
				// ADD LOGIC TO HANDLE REFRESH TOKEN
				// every 24 hours at least
				// if millis is more than 12 hours after last authorization, make new
				currentMillis := time.Now().UnixNano() / 1000000
				refreshMillis := configs.LastAuthMillis + 43200000

				if currentMillis > refreshMillis {
					newAccessToken, err := client.RefreshAccessToken(configs.RefreshToken, configs.MqttServerURI)
					if err != nil {
						log.Error(errors.Wrap(err, "refreshing access token"))
					}
					configs.AccessToken = newAccessToken
					configs.LastAuthMillis = currentMillis
					if err := configs.SaveToFile(); err != nil {
						log.Error(err)
					}
				}
			}
			for i := 0; i < len(configs.WantedHouseholds); i++ {
				HouseholdID := fmt.Sprintf("%v", configs.WantedHouseholds[i])
				states.Groups, states.Players, err = client.GetGroupsAndPlayers(configs.AccessToken, HouseholdID)
				if err != nil {
					log.Error("error")
				}
			}

			for i := 0; i < len(states.Groups); i++ {
				metadata, err := client.GetMetadata(configs.AccessToken, states.Groups[i].GroupId)
				if err == nil {
					states.Container = metadata.Container
					states.CurrentItem = metadata.CurrentItem
					states.NextItem = metadata.NextItem
					if states.Container.Service.Name == "Sonos Radio" {
						states.StreamInfo = metadata.StreamInfo
						states.Container.ImageURL = "https://static.vecteezy.com/system/resources/previews/000/581/923/non_2x/radio-icon-vector-illustration.jpg"
					} else {
						states.StreamInfo = ""
					}

					imageURL := states.Container.ImageURL
					if imageURL == "" {
						imageURL = states.CurrentItem.Track.ImageURL
					}

					report := map[string]interface{}{
						"album":       states.CurrentItem.Track.Album.Name,
						"track":       states.CurrentItem.Track.Name,
						"artist":      states.CurrentItem.Track.Artist,
						"image_url":   imageURL,
						"stream_info": states.StreamInfo,
					}
					adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: states.Groups[i].FimpId}
					oldReportEqualsNewReport := reflect.DeepEqual(oldReport, report)
					if !oldReportEqualsNewReport {
						msg := fimpgo.NewMessage("evt.metadata.report", "media_player", fimpgo.VTypeStrMap, report, nil, nil, nil)
						if err := mqtt.Publish(adr, msg); err != nil {
							log.Error(err)
						}
						oldReport = report
						log.Info("New metadata message sent to fimp")
					}

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

					pbStateValue := client.SetCorrectValue(states.PlaybackState)
					if oldPbStateValue != pbStateValue {
						msg := fimpgo.NewMessage("evt.playback.report", "media_player", fimpgo.VTypeString, pbStateValue, nil, nil, nil)
						if err := mqtt.Publish(adr, msg); err != nil {
							log.Error(err)
						}
						oldPbStateValue = pbStateValue
						log.Info("New playback.report sent to fimp")
					}
					oldPlayModesEqualsNewPlaymodes := reflect.DeepEqual(oldPlayModes, states.PlayModes)
					playmodes := map[string]bool{
						"repeat":     states.PlayModes.Repeat,
						"repeat_one": states.PlayModes.RepeatOne,
						"shuffle":    states.PlayModes.Shuffle,
						"crossfade":  states.PlayModes.Crossfade,
					}
					if !oldPlayModesEqualsNewPlaymodes {
						msg := fimpgo.NewMessage("evt.playbackmode.report", "media_player", fimpgo.VTypeBoolMap, playmodes, nil, nil, nil)
						if err := mqtt.Publish(adr, msg); err != nil {
							log.Error(err)
						}
						oldPlayModes = states.PlayModes
						log.Info("New playbackmode.report sent to fimp")
					}
					if oldVolume != states.Volume {
						msg := fimpgo.NewMessage("evt.volume.report", "media_player", fimpgo.VTypeInt, states.Volume, nil, nil, nil)
						if err := mqtt.Publish(adr, msg); err != nil {
							log.Error(err)
						}
						oldVolume = states.Volume
						log.Info("New volume.report sent to fimp")
					}
					if oldMuted != states.Muted {
						msg := fimpgo.NewMessage("evt.mute.report", "media_player", fimpgo.VTypeBool, states.Muted, nil, nil, nil)
						if err := mqtt.Publish(adr, msg); err != nil {
							log.Error(err)
						}
						oldMuted = states.Muted
						log.Info("New mute.report sent to fimp")
					}

				} else {
					log.Error("This is the one in service.go", err)
				}
			}
			log.Debug("ticker")
			if err := states.SaveToFile(); err != nil {
				log.Error(err)
			}
		}
		appLifecycle.WaitForState(model.AppStateNotConfigured, "main")
	}

}
