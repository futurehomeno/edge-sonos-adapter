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
}

func NewFromFimpRouter(mqt *fimpgo.MqttTransport, appLifecycle *model.Lifecycle, configs *model.Configs, states *model.States) *FromFimpRouter {
	fc := FromFimpRouter{inboundMsgCh: make(fimpgo.MessageCh, 5), mqt: mqt, appLifecycle: appLifecycle, configs: configs, states: states}
	fc.mqt.RegisterChannel("ch1", fc.inboundMsgCh)
	return &fc
}

func (fc *FromFimpRouter) Start() {

	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:dev/rn:%s/ad:1/#", model.ServiceName))
	fc.mqt.Subscribe(fmt.Sprintf("pt:j1/+/rt:ad/rn:%s/ad:1", model.ServiceName))

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
	client := sonos.Client{}
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
			if success {
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.playback.report", "media_player", fimpgo.VTypeStrMap, pbStatus.PlaybackState, nil, nil, newMsg.Payload)
				fc.mqt.Publish(adr, msg)
			}

		case "cmd.playback.get_report":
			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			// send playback status

			success, err := fc.client.GetPlaybackStatus(CorrID, fc.configs.AccessToken)
			if err != nil {
				log.Error(err)
			}
			if success != nil {
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.playback.report", "media_player", fimpgo.VTypeStrMap, fc.states.PlaybackState, nil, nil, newMsg.Payload)
				fc.mqt.Publish(adr, msg)
			}

		case "cmd.mode.set":
			// find groupId from addr(playerId)
			CorrID, err := fc.id.FindGroupFromPlayer(addr, fc.states.Groups)
			if err != nil {
				log.Error(err)
			}

			// get str_map including bool values of repeat, repeatOne, crossfade and shuffle
			val, err := newMsg.Payload.GetStrMapValue()
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
				adr := &fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: model.ServiceName, ResourceAddress: "1", ServiceName: "media_player", ServiceAddress: addr}
				msg := fimpgo.NewMessage("evt.mode.report", "media_player", fimpgo.VTypeStrMap, pbStatus.PlayModes, nil, nil, newMsg.Payload)
				fc.mqt.Publish(adr, msg)
			}

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
				fc.mqt.Publish(adr, msg)
			}

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
				fc.mqt.Publish(adr, msg)
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
				fc.mqt.Publish(adr, msg)
			}
			fc.states.Households, err = fc.client.GetHousehold(fc.configs.AccessToken)
			if err != nil {
				log.Error("error")
			}

			fc.configs.SaveToFile()
			fc.states.SaveToFile()

		case "cmd.auth.logout":
			// exclude all players
			// respond to wanted topic with necessary value(s)
			fc.configs.AccessToken = ""
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
				fc.mqt.Publish(adr, msg)
			}
			fc.states.Households, fc.states.Groups, fc.states.Players = nil, nil, nil

			fc.configs.LoadDefaults()
			fc.states.LoadDefaults()

			val2 := map[string]interface{}{
				"errors":  nil,
				"success": true,
			}
			msg := fimpgo.NewMessage("evt.pd7.response", "vinculum", fimpgo.VTypeObject, val2, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				log.Error("Could not respond to wanted request")
			}
			fc.configs.LoadDefaults()

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
				client.Env = hubInfo.Environment
			} else {
				client.Env = utils.EnvProd
			}

			if client.Env == "beta" {
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
				manifest.Configs[0].UI.Select = nil
			}

			msg := fimpgo.NewMessage("evt.app.manifest_report", model.ServiceName, fimpgo.VTypeObject, manifest, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				fc.mqt.Publish(adr, msg)
			}

		case "cmd.app.get_state":
			msg := fimpgo.NewMessage("evt.app.manifest_report", model.ServiceName, fimpgo.VTypeObject, fc.appLifecycle.GetAllStates(), nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				// if response topic is not set , sending back to default application event topic
				fc.mqt.Publish(adr, msg)
			}

		case "cmd.config.get_extended_report":

			msg := fimpgo.NewMessage("evt.config.extended_report", model.ServiceName, fimpgo.VTypeObject, fc.configs, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
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
			fc.configs.SaveToFile()
			log.Debugf("App reconfigured . New parameters : %v", fc.configs)
			// TODO: This is an example . Add your logic here or remove
			configReport := model.ConfigReport{
				OpStatus: "ok",
				AppState: *fc.appLifecycle.GetAllStates(),
			}
			msg := fimpgo.NewMessage("evt.app.config_report", model.ServiceName, fimpgo.VTypeObject, configReport, nil, nil, newMsg.Payload)
			if err := fc.mqt.RespondToRequest(newMsg.Payload, msg); err != nil {
				fc.mqt.Publish(adr, msg)
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
					fc.mqt.Publish(&adr, msg)
				}
				fc.appLifecycle.SetConfigState(model.ConfigStateConfigured)
			}

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
				fc.configs.SaveToFile()
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
				fc.mqt.Publish(adr, msg)
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
				fc.mqt.Publish(adr, msg)
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
					fc.mqt.Publish(&adr, msg)
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
		fc.configs.SaveToFile()
		fc.states.SaveToFile()

	}

}
