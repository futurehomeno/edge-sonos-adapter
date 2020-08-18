package router

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/thingsplex/sonos/sonos-api"

	"github.com/thingsplex/sonos/handler"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/utils"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/sonos/model"
)

type FromFimpRouter struct {
	inboundMsgCh fimpgo.MessageCh
	mqt          *fimpgo.MqttTransport
	instanceId   string
	appLifecycle *model.Lifecycle
	configs      *model.Configs
	states       *model.States
	pb           *handler.Playback
	vol          *handler.Volume
	client       *sonos.Client
	id           *handler.Id
	mute         *handler.Mute
	fv           *handler.Favorite
	pl           *handler.Playlist
}

func NewFromFimpRouter(mqt *fimpgo.MqttTransport, appLifecycle *model.Lifecycle, configs *model.Configs, states *model.States) *FromFimpRouter {
	fc := FromFimpRouter{inboundMsgCh: make(fimpgo.MessageCh, 5), mqt: mqt, appLifecycle: appLifecycle, configs: configs, states: states}
	fc.mqt.RegisterChannel("ch1", fc.inboundMsgCh)
	return &fc
}

func (fc *FromFimpRouter) Start() {

	if err := fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:dev/rn:%s/ad:1/#", model.ServiceName)); err != nil {
		log.Error(err)
	}
	if err := fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:ad/rn:%s/ad:1", model.ServiceName)); err != nil {
		log.Error(err)
	}

	go func(msgChan fimpgo.MessageCh) {
		for {
			select {
			case newMsg := <-msgChan:
				fc.routeFimpMessage(newMsg)
			}
		}
	}(fc.inboundMsgCh)
}

func (fc *FromFimpRouter) routeFimpMessage(newMsg *fimpgo.Message) {
	log.Debug("New fimp msg")
	addr := strings.Replace(newMsg.Addr.ServiceAddress, "_0", "", 1)
	ns := model.NetworkService{}
	switch newMsg.Payload.Service {

	case "media_player":

		switch newMsg.Payload.Type {
		case "cmd.playback.set":
			// get "play", "pause", "toggle_play_pause", "next_track" or "previous_track"
			val, err := newMsg.Payload.GetStringValue()
			if err != nil {
				log.Error("Ctrl error")
			}

			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			success, err := fc.pb.PlaybackSet(val, CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			pbStatus, err := fc.client.GetPlaybackStatus(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			if pbStatus.PlaybackState == "PLAYBACK_STATE_BUFFERING" {
				pbStatus.PlaybackState = "PLAYBACK_STATE_PLAYING"
			}
			val = fc.client.SetCorrectValue(pbStatus.PlaybackState)
			if success {
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.playback.report", "media_player", fimpgo.VTypeStrMap, val, nil, nil, newMsg.Payload)
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}
			log.Info("New playback.set, ", val)

		case "cmd.playback.get_report":
			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			// send playback status

			pbStatus, err := fc.client.GetPlaybackStatus(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			if pbStatus.PlaybackState == "PLAYBACK_STATE_BUFFERING" {
				pbStatus.PlaybackState = "PLAYBACK_STATE_PLAYING"
			}
			val := fc.client.SetCorrectValue(pbStatus.PlaybackState)
			if pbStatus != nil {
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.playback.report", "media_player", fimpgo.VTypeStrMap, val, nil, nil, newMsg.Payload)
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}
			log.Info("cmd.playback.get_report called")

		case "cmd.playbackmode.set":
			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			// get bool_map including bool values of repeat, repeatOne, crossfade and shuffle
			val, err := newMsg.Payload.GetBoolMapValue()
			if err != nil {
				log.Error("Set mode error")
			}

			success, err := fc.pb.ModeSet(val, CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			pbStatus, err := fc.client.GetPlaybackStatus(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			if success {
				playmodes := map[string]bool{
					"repeat":     pbStatus.PlayModes.Repeat,
					"repeat_one": pbStatus.PlayModes.RepeatOne,
					"shuffle":    pbStatus.PlayModes.Shuffle,
					"crossfade":  pbStatus.PlayModes.Crossfade,
				}
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.playbackmode.report", "media_player", fimpgo.VTypeBoolMap, playmodes, nil, nil, newMsg.Payload)
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}
			log.Info("New playbackmode.report, ", pbStatus.PlayModes)

		case "cmd.playbackmode.get_report":
			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			pbStatus, err := fc.client.GetPlaybackStatus(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
			msg := fimpgo.NewMessage("evt.playbackmode.report", "media_player", fimpgo.VTypeBoolMap, pbStatus.PlayModes, nil, nil, newMsg.Payload)
			if err := fc.mqt.Publish(adr, msg); err != nil {
				log.Error(err)
			}
			log.Info("cmd.playbackmode.get_report called")

		case "cmd.volume.set":
			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			// get int from 0-100 representing new volume in %
			val, err := newMsg.Payload.GetIntValue()
			if err != nil {
				log.Error("Volume error", err)
			}

			success, err := fc.vol.VolumeSet(val, CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			currVolume, err := fc.client.GetVolume(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			if success {
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.volume.report", "media_player", fimpgo.VTypeInt, currVolume.Volume, nil, nil, newMsg.Payload)
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}
			log.Info("New volume set, ", val)

		case "cmd.volume.get_report":
			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}
			currVolume, err := fc.client.GetVolume(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
			msg := fimpgo.NewMessage("evt.volume.report", "media_player", fimpgo.VTypeInt, currVolume.Volume, nil, nil, newMsg.Payload)
			if err := fc.mqt.Publish(adr, msg); err != nil {
				log.Error(err)
			}
			log.Info("cmd.volume.get_report called")

		case "cmd.mute.set":
			// find groupId fom addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			// get bool value
			val, err := newMsg.Payload.GetBoolValue()
			if err != nil {
				log.Error("Volume error", err)
			}

			success, err := fc.mute.MuteSet(val, CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			currVolume, err := fc.client.GetVolume(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			if success {
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.mute.report", "media_player", fimpgo.VTypeStrMap, currVolume.Muted, nil, nil, newMsg.Payload)
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}
			log.Info("New mute set, ", val)

		case "cmd.mute.get_report":
			// find groupId fom addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}
			currVolume, err := fc.client.GetVolume(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
			msg := fimpgo.NewMessage("evt.mute.report", "media_player", fimpgo.VTypeStrMap, currVolume.Muted, nil, nil, newMsg.Payload)
			if err := fc.mqt.Publish(adr, msg); err != nil {
				log.Error(err)
			}
			log.Info("cmd.mute.get_report called")

		case "cmd.favorites.get_report":
			var err error
			HouseholdID := fmt.Sprintf("%v", fc.configs.WantedHouseholds[0])
			fc.states.Favorites, err = fc.client.GetFavorites(fc.configs.AccessToken, HouseholdID)
			// fc.states.Favorites, err = fc.fv.GetFavorites(fc.configs.AccessToken, fc.states.Households[0].ID)
			if err != nil {
				log.Error(err)
			}
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
			msg := fimpgo.NewMessage("evt.favorites.report", "media_player", fimpgo.VTypeObject, fc.states.Favorites, nil, nil, newMsg.Payload)
			fc.mqt.Publish(adr, msg)
			log.Info("cmd.favorites.get_report called")

		case "cmd.favorites.set":
			val, err := newMsg.Payload.GetStringValue()
			if err != nil {
				log.Error(err)
			}
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}
			log.Debug("song id: ", val)
			success, err := fc.fv.FavoriteSet(val, CorrID, fc.configs.AccessToken)
			if success {
				metadata, err := fc.client.GetMetadata(fc.configs.AccessToken, CorrID)
				if err == nil {
					fc.states.Container = metadata.Container
					fc.states.CurrentItem = metadata.CurrentItem
					fc.states.NextItem = metadata.NextItem
					if fc.states.Container.Service.Name == "Sonos Radio" {
						fc.states.StreamInfo = metadata.StreamInfo
						fc.states.Container.ImageURL = "https://static.vecteezy.com/system/resources/previews/000/581/923/non_2x/radio-icon-vector-illustration.jpg"
					} else {
						fc.states.StreamInfo = ""
					}

					imageURL := fc.states.Container.ImageURL
					if imageURL == "" {
						imageURL = fc.states.CurrentItem.Track.ImageURL
					}

					report := map[string]interface{}{
						"album":       fc.states.CurrentItem.Track.Album.Name,
						"track":       fc.states.CurrentItem.Track.Name,
						"artist":      fc.states.CurrentItem.Track.Artist,
						"image_url":   imageURL,
						"stream_info": fc.states.StreamInfo,
					}
					adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}

					msg := fimpgo.NewMessage("evt.metadata.report", "media_player", fimpgo.VTypeStrMap, report, nil, nil, nil)
					if err := fc.mqt.Publish(adr, msg); err != nil {
						log.Error(err)
					}
					log.Info("New metadata message sent to fimp")
				}
			}

		case "cmd.playlists.get_report":
			var err error
			HouseholdID := fmt.Sprintf("%v", fc.configs.WantedHouseholds[0])
			fc.states.Playlists, err = fc.client.GetPlaylists(fc.configs.AccessToken, HouseholdID)
			if err != nil {
				log.Error(err)
			}
			adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
			msg := fimpgo.NewMessage("evt.playlists.report", "media_player", fimpgo.VTypeObject, fc.states.Playlists, nil, nil, newMsg.Payload)
			fc.mqt.Publish(adr, msg)
			log.Info("cmd.playlists.get_report called")

		case "cmd.playlists.set":
			val, err := newMsg.Payload.GetStringValue()
			if err != nil {
				log.Error(err)
			}
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}
			log.Debug("playlist id: ", val)
			success, err := fc.pl.PlaylistSet(val, CorrID, fc.configs.AccessToken)
			if success {
				metadata, err := fc.client.GetMetadata(fc.configs.AccessToken, CorrID)
				if err == nil {
					fc.states.Container = metadata.Container
					fc.states.CurrentItem = metadata.CurrentItem
					fc.states.NextItem = metadata.NextItem
					if fc.states.Container.Service.Name == "Sonos Radio" {
						fc.states.StreamInfo = metadata.StreamInfo
						fc.states.Container.ImageURL = "https://static.vecteezy.com/system/resources/previews/000/581/923/non_2x/radio-icon-vector-illustration.jpg"
					} else {
						fc.states.StreamInfo = ""
					}

					imageURL := fc.states.Container.ImageURL
					if imageURL == "" {
						imageURL = fc.states.CurrentItem.Track.ImageURL
					}

					report := map[string]interface{}{
						"album":       fc.states.CurrentItem.Track.Album.Name,
						"track":       fc.states.CurrentItem.Track.Name,
						"artist":      fc.states.CurrentItem.Track.Artist,
						"image_url":   imageURL,
						"stream_info": fc.states.StreamInfo,
					}
					adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}

					msg := fimpgo.NewMessage("evt.metadata.report", "media_player", fimpgo.VTypeStrMap, report, nil, nil, nil)
					if err := fc.mqt.Publish(adr, msg); err != nil {
						log.Error(err)
					}
					log.Info("New metadata message sent to fimp")
				}
			}
		}

	case model.ServiceName:
		adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: model.ServiceName, ResourceAddress: "1"}
		switch newMsg.Payload.Type {

		case "cmd.auth.set_tokens":
			authReq := model.SetTokens{}
			err := newMsg.Payload.GetObjectValue(&authReq)
			if err != nil {
				log.Error("Incorrect login message ")
				return
			}
			status := model.AuthStatus{
				Status:    model.AuthStateAuthenticated,
				ErrorText: "",
				ErrorCode: "",
			}
			if authReq.AccessToken != "" && authReq.RefreshToken != "" {
				fc.configs.AccessToken = authReq.AccessToken
				fc.configs.LastAuthMillis = time.Now().UnixNano() / 1000000
				fc.configs.RefreshToken = authReq.RefreshToken
				fc.configs.ExpiresIn = authReq.ExpiresIn
				fc.appLifecycle.SetAuthState(model.AuthStateAuthenticated)
				fc.appLifecycle.SetConnectionState(model.ConnStateConnected)
			} else {
				status.Status = "ERROR"
				status.ErrorText = "Empty username or password"
			}
			msg := fimpgo.NewMessage("evt.auth.status_report", model.ServiceName, fimpgo.VTypeObject, status, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}
			fc.states.Households, err = fc.client.GetHousehold(fc.configs.AccessToken)
			if err != nil {
				log.Error("error")
			}
			log.Info("New tokens set successfully")
			if err := fc.configs.SaveToFile(); err != nil {
				log.Error(err)
			}
			if err := fc.states.SaveToFile(); err != nil {
				log.Error(err)
			}

		case "cmd.auth.logout":
			// exclude all players
			// respond to wanted topic with necessary value(s)
			fc.configs.AccessToken = ""
			fc.configs.WantedHouseholds = nil
			fc.configs.LastAuthMillis = 0
			fc.appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
			fc.appLifecycle.SetAuthState(model.AuthStateNotAuthenticated)
			fc.appLifecycle.SetConnectionState(model.ConnStateDisconnected)

			for i := 0; i < len(fc.states.Players); i++ {
				playerID := fc.states.Players[i].FimpId
				val := map[string]interface{}{
					"address": playerID,
				}
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "sonos", ResourceAddress: "1"}
				msg := fimpgo.NewMessage("evt.thing.exclusion_report", "sonos", fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}
			fc.states.Households, fc.states.Groups, fc.states.Players = nil, nil, nil

			if err := fc.configs.LoadDefaults(); err != nil {
				log.Error(err)
			}
			if err := fc.states.LoadDefaults(); err != nil {
				log.Error(err)
			}

			val2 := map[string]interface{}{
				"errors":  nil,
				"success": true,
			}
			msg := fimpgo.NewMessage("evt.pd7.response", "vinculum", fimpgo.VTypeObject, val2, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				log.Error("Could not respond to wanted request")
			}
			log.Info("Logged out successfully")
			if err := fc.configs.LoadDefaults(); err != nil {
				log.Error(err)
			}

		case "cmd.app.get_manifest":
			mode, err := newMsg.Payload.GetStringValue()
			if err != nil {
				log.Error("Incorrect request format ")
				return
			}
			manifest := model.NewManifest()
			err = manifest.LoadFromFile(filepath.Join(fc.configs.GetDefaultDir(), "app-manifest.json"))
			if err != nil {
				log.Error("Failed to load manifest file .Error :", err.Error())
				return
			}
			hubInfo, err := utils.NewHubUtils().GetHubInfo()

			if err == nil && hubInfo != nil {
				fc.configs.Env = hubInfo.Environment
			} else {
				fc.configs.Env = utils.EnvProd
			}

			if fc.configs.Env == "beta" {
				manifest.Auth.RedirectURL = "https://app-static-beta.futurehome.io/playground_oauth_callback"
				manifest.Auth.AuthEndpoint = "https://partners-beta.futurehome.io/api/edge/proxy/auth-code"
			} else {
				manifest.Auth.RedirectURL = "https://app-static.futurehome.io/playground_oauth_callback"
				manifest.Auth.AuthEndpoint = "https://partners.futurehome.io/api/edge/proxy/auth-code"
			}
			if mode == "manifest_state" {
				manifest.AppState = *fc.appLifecycle.GetAllStates()
				manifest.ConfigState = fc.configs
			}
			if fc.configs.IsAuthenticated() {
				var householdSelect []interface{}
				manifest.Configs[0].ValT = "str_map"
				manifest.Configs[0].UI.Type = "list_checkbox"
				for i := 0; i < len(fc.states.Households); i++ {
					HouseholdID := fmt.Sprintf("%v", fc.states.Households[i].ID)
					_, players, err := fc.client.GetGroupsAndPlayers(fc.configs.AccessToken, HouseholdID)
					if err != nil {
						log.Error("error: ", err)
					}
					numPlayers := len(players)
					if numPlayers > 1 {
						label := fmt.Sprintf("System with %d devices", numPlayers)
						householdSelect = append(householdSelect, map[string]interface{}{"val": fc.states.Households[i].ID, "label": map[string]interface{}{"en": label}})
					} else {
						label := fmt.Sprintf("System with %d device", numPlayers)
						householdSelect = append(householdSelect, map[string]interface{}{"val": fc.states.Households[i].ID, "label": map[string]interface{}{"en": label}})
					}
				}
				manifest.Configs[0].UI.Select = householdSelect
			} else {
				manifest.Configs[0].ValT = "string"
				manifest.Configs[0].UI.Type = "input_readonly"
				var val model.Value
				val.Default = "You need to login first"
				manifest.Configs[0].Val = val
				// manifest.Configs[0].UI.Select = nil
			}

			msg := fimpgo.NewMessage("evt.app.manifest_report", model.ServiceName, fimpgo.VTypeObject, manifest, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}

		case "cmd.app.get_state":
			msg := fimpgo.NewMessage("evt.app.manifest_report", model.ServiceName, fimpgo.VTypeObject, fc.appLifecycle.GetAllStates(), nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}

		case "cmd.config.get_extended_report":

			msg := fimpgo.NewMessage("evt.config.extended_report", model.ServiceName, fimpgo.VTypeObject, fc.configs, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}

		case "cmd.config.extended_set":
			conf := model.Configs{}
			err := newMsg.Payload.GetObjectValue(&conf)
			if err != nil {
				// TODO: This is an example . Add your logic here or remove
				log.Error("Can't parse configuration object")
				return
			}
			fc.configs.WantedHouseholds = conf.WantedHouseholds
			if err := fc.configs.SaveToFile(); err != nil {
				log.Error(err)
			}
			log.Debugf("App reconfigured . New parameters : %v", fc.configs)
			// TODO: This is an example . Add your logic here or remove
			configReport := model.ConfigReport{
				OpStatus: "ok",
				AppState: *fc.appLifecycle.GetAllStates(),
			}
			msg := fimpgo.NewMessage("evt.app.config_report", model.ServiceName, fimpgo.VTypeObject, configReport, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}

			for i := 0; i < len(fc.configs.WantedHouseholds); i++ {
				HouseholdID := fmt.Sprintf("%v", fc.configs.WantedHouseholds[i])
				fc.states.Groups, fc.states.Players, err = fc.client.GetGroupsAndPlayers(fc.configs.AccessToken, HouseholdID)

				if err != nil {
					log.Error("error")
				}
				for i := 0; i < len(fc.states.Players); i++ {
					inclReport := ns.MakeInclusionReport(fc.states.Players[i])

					msg := fimpgo.NewMessage("evt.thing.inclusion_report", "sonos", fimpgo.VTypeObject, inclReport, nil, nil, nil)
					adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "sonos", ResourceAddress: "1"}
					if err := fc.mqt.Publish(&adr, msg); err != nil {
						log.Error(err)
					}
				}
				fc.appLifecycle.SetConfigState(model.ConfigStateConfigured)
			}
			log.Info("Wanted households updated, new wanted households: ", fc.configs.WantedHouseholds)

		case "cmd.log.set_level":
			// Configure log level
			level, err := newMsg.Payload.GetStringValue()
			if err != nil {
				return
			}
			logLevel, err := log.ParseLevel(level)
			if err == nil {
				log.SetLevel(logLevel)
				fc.configs.LogLevel = level
				if err := fc.configs.SaveToFile(); err != nil {
					log.Error(err)
				}
			}
			log.Info("Log level updated to = ", logLevel)

		case "cmd.system.reconnect":
			// This is optional operation.
			fc.appLifecycle.PublishEvent(model.EventConfigured, "from-fimp-router", nil)
			//val := map[string]string{"status":status,"error":errStr}
			val := model.ButtonActionResponse{
				Operation:       "cmd.system.reconnect",
				OperationStatus: "ok",
				Next:            "config",
				ErrorCode:       "",
				ErrorText:       "",
			}
			msg := fimpgo.NewMessage("evt.app.config_action_report", model.ServiceName, fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}

		case "cmd.app.factory_reset":
			val := model.ButtonActionResponse{
				Operation:       "cmd.app.factory_reset",
				OperationStatus: "ok",
				Next:            "config",
				ErrorCode:       "",
				ErrorText:       "",
			}
			fc.appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
			fc.appLifecycle.SetAppState(model.AppStateNotConfigured, nil)
			fc.appLifecycle.SetAuthState(model.AuthStateNotAuthenticated)
			msg := fimpgo.NewMessage("evt.app.config_action_report", model.ServiceName, fimpgo.VTypeObject, val, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				if err := fc.mqt.Publish(adr, msg); err != nil {
					log.Error(err)
				}
			}

		case "cmd.network.get_all_nodes":
			// TODO: This is an example . Add your logic here or remove
		case "cmd.thing.get_inclusion_report":
			nodeId, _ := newMsg.Payload.GetStringValue()
			for i := 0; i < len(fc.states.Players); i++ {
				if nodeId == fc.states.Players[i].FimpId {
					Player := fc.states.Players[i]
					inclReport := ns.MakeInclusionReport(Player)

					msg := fimpgo.NewMessage("evt.thing.inclusion_report", "sonos", fimpgo.VTypeObject, inclReport, nil, nil, nil)
					adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeAdapter, ResourceName: "sonos", ResourceAddress: "1"}
					if err := fc.mqt.Publish(&adr, msg); err != nil {
						log.Error(err)
					}
				}
			}
		case "cmd.thing.inclusion":
			//flag , _ := newMsg.Payload.GetBoolValue()
			// TODO: This is an example . Add your logic here or remove
		case "cmd.thing.delete":
			// remove device from network
			val, err := newMsg.Payload.GetStrMapValue()
			if err != nil {
				log.Error("Wrong msg format")
				return
			}
			deviceId, ok := val["address"]
			if ok {
				// TODO: This is an example . Add your logic here or remove
				log.Info(deviceId)
			} else {
				log.Error("Incorrect address")

			}
		}
		if err := fc.configs.SaveToFile(); err != nil {
			log.Error(err)
		}
		if err := fc.states.SaveToFile(); err != nil {
			log.Error(err)
		}

	}

}
