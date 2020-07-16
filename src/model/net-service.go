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
func (ns *NetworkService) MakeInclusionReport(Group sonos.Group) fimptype.ThingInclusionReport {
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
		MsgType:   "cmd.mode.set",
		ValueType: "str_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.mode.get_report",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.mode.report",
		ValueType: "str_map",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.volume.set",
		ValueType: "int",
		Version:   "1",
	}, {
		Type:      "in",
		MsgType:   "cmd.volume.get",
		ValueType: "null",
		Version:   "1",
	}, {
		Type:      "out",
		MsgType:   "evt.volume.report",
		ValueType: "str_map",
		Version:   "1",
	}}

	mediaPlayerService := fimptype.Service{
		Name:    "media_player",
		Alias:   "media_player",
		Address: "/rt:dev/rn:sonos/ad:1/sv:media_player/ad:",
		Enabled: true,
		Groups:  []string{"ch_0"},
		Props: map[string]interface{}{
			"sup_playback": []string{"play", "pause", "togglePlayPause", "skipToNextTrack", "skipToPreviousTrack"},
			"sup_modes":    []string{"repeat", "repeatOne", "shuffle", "crossfade"},
		},
		Interfaces: mediaPlayerInterfaces,
	}

	groupID := Group.OnlyGroupId
	manufacturer = "sonos"
	name = Group.Name
	serviceAddress := fmt.Sprintf("%s", groupID)
	mediaPlayerService.Address = mediaPlayerService.Address + serviceAddress
	services = append(services, mediaPlayerService)
	deviceAddr = fmt.Sprintf("%s", groupID)
	powerSource := "ac"

	inclReport := fimptype.ThingInclusionReport{
		IntegrationId:     "",
		Address:           deviceAddr,
		Type:              "",
		ProductHash:       manufacturer,
		CommTechnology:    "wifi",
		ProductName:       name,
		ManufacturerId:    manufacturer,
		DeviceId:          groupID,
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
