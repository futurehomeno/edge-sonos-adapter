package handler

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Playback attributes
type Playback struct {
	status string
}

// PlaybackSet sends request to Sonos to play, pause or skip
func (pb *Playback) PlaybackSet(val string, id string) {
	url := fmt.Sprintf("%s%s%s%s", "https://api.sonos.com/groups/", id, "/playback/", val)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Error("Error when setting playback: ", err)
		return
	}
	log.Debug("New request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on playbackSet: ", err)
	}
	log.Debug("playbackSet first resp: ", resp)

}

// GetPlaybackStatus gets playback status
func (pb *Playback) GetPlaybackStatus(id string) {
	url := fmt.Sprintf("%s%s%s", "https://api.sonos.com/groups/", id, "/playback")
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Error("Error when getting playback status: ", err)
		return
	}
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on getPlaybackStatus: ", err)
	}
	log.Debug("playbackStatus first resp: ", resp)

}
