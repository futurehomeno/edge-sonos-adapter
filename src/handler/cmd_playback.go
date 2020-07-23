package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/thingsplex/sonos/model"

	log "github.com/sirupsen/logrus"
)

// Playback attributes
type Playback struct {
	states *model.States
}

// PlaybackSet sends request to Sonos to play, pause or skip
func (pb *Playback) PlaybackSet(val string, id string, accessToken string) (bool, error) {
	// change toggle_play_pause to togglePlayPause, next_track to skipToNextTrack and previous_track to skipToPreviousTrack
	log.Debug("THIS IS THE VALUE BEFORE: ", val)
	if val == "toggle_play_pause" {
		val = "togglePlayPause"
	} else if val == "next_track" {
		val = "skipToNextTrack"
	} else if val == "previous_track" {
		val = "skipToPreviousTrack"
	}
	log.Debug("THIS IS THE VALUE AFTER: ", val)

	url := fmt.Sprintf("%s%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback/", val)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Error("Error when setting playback: ", err)
		return false, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on PlaybackSet: ", err)
		return false, err
	}

	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return false, fmt.Errorf("%s%s", "Bad HTTP return code ", strconv.Itoa(resp.StatusCode))
	}

	return true, nil
}

// ModeSet sets new mode
func (pb *Playback) ModeSet(val map[string]string, id string, accessToken string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback/playMode")
	// if one of the modes is repeat_one, change to repeatOne

	log.Debug("THIS IS THE VALUE BEFORE: ", val)
	if _, ok := val["repeat_one"]; ok {
		val["repeatOne"] = val["repeat_one"]
		delete(val, "repeat_one")
	}
	log.Debug("THIS IS THE VALUE AFTER: ", val)

	mode := map[string]interface{}{
		"playModes": val,
	}

	bytesRepresentation, err := json.Marshal(mode)
	if err != nil {
		log.Error(err)
		return false, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Error("Error when setting mode: ", err)
		return false, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on ModeSet: ", err)
	}
	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return false, fmt.Errorf("%s%s", "Bad HTTP return code ", strconv.Itoa(resp.StatusCode))
	}

	return true, nil

}
