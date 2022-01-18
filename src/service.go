package main

import (
	"flag"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/thoas/go-funk"

	"github.com/futurehomeno/edge-sonos-adapter/model"
	router "github.com/futurehomeno/edge-sonos-adapter/router"
	"github.com/futurehomeno/edge-sonos-adapter/sonos-api"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
)

const defaultTickerInterval = 300

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

	client := sonos.NewClient(configs.Env, configs.AccessToken, configs.RefreshToken)
	client.UpdateAuthParameters(configs.MqttServerURI)
	edgeapp.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting sonos----------------")
	log.Info("Work directory : ", configs.WorkDir)
	appLifecycle.SetAppState(model.AppStateNotConfigured, nil)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	defer mqtt.Stop()

	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()

	fimpRouter := router.NewFromFimpRouter(mqtt, appLifecycle, configs, states, client)
	fimpRouter.Start()

	appLifecycle.SetConnectionState(model.ConnStateDisconnected)

	// Checking internet connection
	systemCheck := edgeapp.NewSystemCheck()
	err = systemCheck.WaitForInternet(time.Second * 60)
	if err != nil {
		log.Error("<main> Internet is not available, the adapter might not work.")
	}

	if configs.IsAuthenticated() && err == nil {
		appLifecycle.SetAuthState(model.AuthStateAuthenticated)
		for i := 0; i < 5; i++ {
			states.Households, err = client.GetHousehold()
			if err == nil {
				appLifecycle.SetConnectionState(model.ConnStateConnected)
				break
			} else {
				log.Error("<main> Can't get household.Retrying.... Err:", err.Error())
				time.Sleep(time.Second * 10)
				appLifecycle.SetConnectionState(model.ConnStateDisconnected)
			}
		}
	} else {
		appLifecycle.SetAuthState(model.AuthStateNotAuthenticated)
	}

	if configs.IsConfigured() && err == nil {
		appLifecycle.SetConfigState(model.ConfigStateConfigured)
		appLifecycle.SetAppState(model.AppStateRunning, nil)
	} else {
		appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
	}

	// Set tickerInterval variable to configured value. If this value has never been configured it will be 0, set to default 300.
	router.TickerInterval = configs.PollRate
	if router.TickerInterval == 0 {
		router.TickerInterval = defaultTickerInterval
	}

	initialTickerInterval := router.TickerInterval

	for {
		appLifecycle.WaitForState("main", model.AppStateRunning)
		log.Info("<main>Starting update loop")
		LoadStates(configs, client, states, err, mqtt)
		ticker := time.NewTicker(time.Duration(router.TickerInterval) * time.Second)

		for range ticker.C {
			if initialTickerInterval != router.TickerInterval {
				log.Debug("New ticker interval value. Restarting ticker.")
				initialTickerInterval = router.TickerInterval

				break // restart ticker
			}

			if appLifecycle.AppState() != model.AppStateRunning {
				break
			}
			states = LoadStates(configs, client, states, err, mqtt)
		}
		ticker.Stop()
	}

}

func LoadStates(configs *model.Configs, client *sonos.Client, states *model.States, err error, mqtt *fimpgo.MqttTransport) *model.States {

	if configs.AccessToken != "" && configs.AccessToken != "access_token" {
		// ADD LOGIC TO HANDLE REFRESH TOKEN
		// every 24 hours at least
		// if millis is more than 12 hours after last authorization, make new
		currentMillis := time.Now().UnixNano() / 1000000
		refreshMillis := configs.LastAuthMillis + 43200000

		if currentMillis > refreshMillis {
			log.Debug("<main> Access token expired , requesting new token")
			newAccessToken, err := client.RefreshAccessToken(configs.RefreshToken)
			if err != nil {
				log.Error("<main> Can't refresh token")
				log.Error(errors.Wrap(err, "refreshing access token"))
				return states
			} else if newAccessToken != "" {
				configs.AccessToken = newAccessToken
				configs.LastAuthMillis = currentMillis
				client.SetTokens(configs.AccessToken, configs.RefreshToken)
				if err := configs.SaveToFile(); err != nil {
					log.Error("<main> Can't save configurations . Err:", err)
				}
			}
		}
	}
	oldGroups := states.Groups

	for i := 0; i < len(configs.WantedHouseholds); i++ {
		HouseholdID := fmt.Sprintf("%v", configs.WantedHouseholds[i])

		states.Groups, states.Players, err = client.GetGroupsAndPlayers(HouseholdID)
		if err != nil {
			log.Error("error")
		}
	}

	for i, group := range states.Groups {
		if len(oldGroups) != 0 && len(oldGroups) == len(states.Groups) {
			states.Groups[i].OldReport = oldGroups[i].OldReport
			states.Groups[i].OldPbStateValue = oldGroups[i].OldPbStateValue
			states.Groups[i].OldPlayModes = oldGroups[i].OldPlayModes
			states.Groups[i].OldVolume = oldGroups[i].OldVolume
			states.Groups[i].OldMuted = oldGroups[i].OldMuted
		} else if len(oldGroups) != len(states.Groups) {
			return states
		}
		group = states.Groups[i]

		metadata, err := client.GetMetadata(group.GroupId)
		if err == nil {
			var imageURL string
			states.Container = metadata.Container
			states.CurrentItem = metadata.CurrentItem
			states.NextItem = metadata.NextItem
			if states.Container.Service.Name == "Sonos Radio" {
				if metadata.StreamInfo != "" {
					states.StreamInfo = metadata.StreamInfo
				} else {
					states.StreamInfo = ""
				}
				states.CurrentItem = sonos.CurrentItem{}
				if metadata.Container.Name != "" {
					states.CurrentItem.Track.Artist.Name = metadata.Container.Name
				}

				imageURL = ""
				states.NextItem = sonos.NextItem{}
				states.IsRadio = true
			} else {
				states.StreamInfo = ""
				imageURL = states.CurrentItem.Track.ImageURL
				states.IsRadio = false
				if imageURL == "" {
					imageURL = states.Container.ImageURL
				}
			}

			// imageURL := states.Container.ImageURL
			// if imageURL == "" {
			// 	imageURL = states.CurrentItem.Track.ImageURL
			// }

			report := map[string]interface{}{
				"album":       states.CurrentItem.Track.Album.Name,
				"track":       states.CurrentItem.Track.Name,
				"artist":      states.CurrentItem.Track.Artist.Name,
				"image_url":   imageURL,
				"stream_info": states.StreamInfo,
				"is_radio":    states.IsRadio,
			}

			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: group.FimpId}
			oldReportEqualsNewReport := reflect.DeepEqual(group.OldReport, report)
			if !oldReportEqualsNewReport {
				msg := fimpgo.NewMessage("evt.metadata.report", "media_player", fimpgo.VTypeObject, report, nil, nil, nil)
				if err := mqtt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
				group.OldReport = report
				log.Info("New metadata message sent to fimp")
			}

			pbState, err := client.PlaybackGetStatus(group.GroupId)
			if err != nil {
				break
			}
			volume, err := client.VolumeGet(group.GroupId)
			if err != nil {
				break
			}

			states.PlaybackState = pbState.PlaybackState
			states.PlayModes.Crossfade = pbState.PlayModes.Crossfade
			states.PlayModes.Repeat = pbState.PlayModes.Repeat
			states.PlayModes.RepeatOne = pbState.PlayModes.RepeatOne
			states.PlayModes.Shuffle = pbState.PlayModes.Shuffle
			states.Volume = volume.Volume
			states.Muted = volume.Muted
			states.Fixed = volume.Fixed

			pbStateValue := client.SetCorrectValue(states.PlaybackState)
			if group.OldPbStateValue != pbStateValue {
				msg := fimpgo.NewMessage("evt.playback.report", "media_player", fimpgo.VTypeString, pbStateValue, nil, nil, nil)
				if err := mqtt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
				group.OldPbStateValue = pbStateValue
				log.Info("New playback.report sent to fimp")
			}
			oldPlayModesEqualsNewPlaymodes := reflect.DeepEqual(group.OldPlayModes, states.PlayModes)
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
				group.OldPlayModes = states.PlayModes
				log.Info("New playbackmode.report sent to fimp")
			}
			if group.OldVolume != states.Volume {
				msg := fimpgo.NewMessage("evt.volume.report", "media_player", fimpgo.VTypeInt, states.Volume, nil, nil, nil)
				if err := mqtt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
				group.OldVolume = states.Volume
				log.Info("New volume.report sent to fimp")
			}
			if group.OldMuted != states.Muted {
				msg := fimpgo.NewMessage("evt.mute.report", "media_player", fimpgo.VTypeBool, states.Muted, nil, nil, nil)
				if err := mqtt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
				group.OldMuted = states.Muted
				log.Info("New mute.report sent to fimp")
			}
			states.Groups[i] = group
		} else {
			log.Error("This is the one in service.go", err)
		}
	}
	log.Debug("ticker")
	return states
}
