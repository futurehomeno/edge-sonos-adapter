package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/sonos/model"
	"github.com/thingsplex/sonos/router"
	"github.com/thingsplex/sonos/sonos-api"
	"github.com/thingsplex/sonos/utils"
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

	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
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
		ticker := time.NewTicker(time.Duration(60) * time.Second)
		for ; true; <-ticker.C {
			// states.LoadFromFile()
			// configs.LoadFromFile()
			for i := 0; i < len(states.Groups); i++ {
				metadata, err := client.GetMetadata(configs.AccessToken, states.Groups[i].GroupId)
				if err != nil {
					log.Error(err)
				}
				pbState, err := client.GetPlaybackStatus(states.Groups[i].GroupId, configs.AccessToken)
				if err != nil {
					log.Error(err)
				}
				volume, err := client.GetVolume(states.Groups[i].GroupId, configs.AccessToken)
				if err != nil {
					log.Error(err)
				}
				states.Container = metadata.Container
				states.CurrentItem = metadata.CurrentItem
				states.NextItem = metadata.NextItem
				states.PlaybackState = pbState.PlaybackState
				states.PlayModes = pbState.PlayModes
				states.Volume = volume.Volume
				states.Muted = volume.Muted
				states.Fixed = volume.Fixed

				report := map[string]interface{}{
					"album":     states.CurrentItem.Track.Album.Name,
					"track":     states.CurrentItem.Track.Name,
					"artist":    states.CurrentItem.Track.Artist,
					"image_url": states.Container.ImageURL,
				}
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: states.Groups[i].FimpId}
				msg := fimpgo.NewMessage("evt.metadata.report", "media_player", fimpgo.VTypeStrMap, report, nil, nil, nil)
				mqtt.Publish(adr, msg)

				log.Debug(metadata.Container.ImageURL)
				log.Debug(states.Container.ImageURL)
			}
			log.Debug("ticker")
			states.SaveToFile()
		}
		appLifecycle.WaitForState(model.AppStateNotConfigured, "main")
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
