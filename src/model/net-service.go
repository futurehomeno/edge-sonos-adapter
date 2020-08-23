package model

import (
	"fmt"

	"github.com/futurehomeno/fimpgo/fimptype"
	"github.com/thingsplex/sonos/sonos-api"
)

// NetworkService is for export
type NetworkService struct {
}

// MakeInclusionReport makes inclusion report for player with id given in parameter
func (ns *NetworkService) MakeInclusionReport(Player sonos.Player) fimptype.ThingInclusionReport {
	// var err error

	var name, manufacturer string
	var deviceAddr string
	services := []fimptype.Service{}

	mediaPlayerInterfaces := []fimptype.Interface{{
		Type:      "in",
		MsgType:   "cmd.playback.set",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.playback.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.playback.report",
		ValueType: "str_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.playbackmode.set",
		ValueType: "str_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.playbackmode.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.playbackmode.report",
		ValueType: "str_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.volume.set",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.volume.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.volume.report",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.mute.set",
		ValueType: "bool",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.mute.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.mute.report",
		ValueType: "bool",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.metadata.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.metadata.report",
		ValueType: "str_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.favorites.set",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.favorites.get_report",
		ValueType: "null",
		Version:   "1",
	},{
		Type:      "out",
		MsgType:   "evt.favorites.report",
		ValueType: "object",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.playlists.set",
		ValueType: "string",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.playlists.get_report",
		ValueType: "null",
		Version:   "1",
	},{
		Type:      "out",
		MsgType:   "evt.playlists.report",
		ValueType: "object",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.audioclip.play",
		ValueType: "object",
		Version:   "1",
	}}

	mediaPlayerService := fimptype.Service{
		Name:    "media_player",
		Alias:   "media_player",
		Address: "/rt:dev/rn:sonos/ad:1/sv:media_player/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_playback": []string{"play", "pause", "toggle_play_pause", "next_track", "previous_track"},
			"sup_modes":    []string{"repeat", "repeat_one", "shuffle", "crossfade"},
			"sup_metadata": []string{"album", "track", "artist", "image_url"},
		},
		Interfaces: mediaPlayerInterfaces,
	}

	playerID := Player.FimpId
	manufacturer = "sonos"
	name = Player.Name
	serviceAddress := fmt.Sprintf("%s", playerID)
	mediaPlayerService.Address = mediaPlayerService.Address + serviceAddress
	services = append(services, mediaPlayerService)
	deviceAddr = fmt.Sprintf("%s", playerID)
	powerSource := "ac"

	inclReport := fimptype.ThingInclusionReport{
		IntegrationId:     "",
		Address:           deviceAddr,
		Type:              "",
		ProductHash:       manufacturer,
		CommTechnology:    "wifi",
		ProductName:       name,
		ManufacturerId:    manufacturer,
		DeviceId:          playerID,
		HwVersion:         "1",
		SwVersion:         "1",
		PowerSource:       powerSource,
		WakeUpInterval:    "-1",
		Security:          "",
		Tags:              nil,
		Groups:            []string{"ch_0"},
		PropSets:          nil,
		TechSpecificProps: nil,
		Services:          services,
	}

	return inclReport
}
