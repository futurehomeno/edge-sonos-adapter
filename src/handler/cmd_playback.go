package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/thingsplex/sonos/model"

	log "github.com/sirupsen/logrus"
)

// Playback attributes
type Playback struct {
	states *model.States
}

// PlaybackSet sends request to Sonos to play, pause or skip
func (pb *Playback) PlaybackSet(val string, id string, accessToken string) error {
	url := fmt.Sprintf("%s%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback/", val)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Error("Error when setting playback: ", err)
		return err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on PlaybackSet: ", err)
	}
	log.Debug("playbackSet first resp: ", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetPlaybackStatus: ", err)
		return err
	}
	log.Debug(body)
	return nil
}

// GetPlaybackStatus gets playback status
func (pb *Playback) GetPlaybackStatus(id string, accessToken string) *Playback {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Error when getting playback status: ", err)
		return nil
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetPlaybackStatus: ", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetPlaybackStatus: ", err)
		return nil
	}
	err = json.Unmarshal(body, &pb)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil
	}
	return pb
}

// ModeSet sets new mode
func (pb *Playback) ModeSet(val map[string]string, id string, accessToken string) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback/playMode")

	mode := map[string]interface{}{
		"playModes": val,
	}
	bytesRepresentation, err := json.Marshal(mode)
	if err != nil {
		log.Error(err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Error("Error when setting mode: ", err)
		return
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on ModeSet: ", err)
	}
	log.Debug("modeSet first resp: ", resp)
}
