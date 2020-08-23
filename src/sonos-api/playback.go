package sonos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type PlaybackStatusResponse struct {
	PlaybackState  string `json:"playbackState"`
	QueueVersion   string `json:"queueVersion"`
	ItemID         string `json:"itemId"`
	PositionMillis int    `json:"positionMillis"`
	PlayModes      struct {
		Repeat    bool `json:"repeat"`
		RepeatOne bool `json:"repeatOne"`
		Crossfade bool `json:"crossfade"`
		Shuffle   bool `json:"shuffle"`
	} `json:"playModes"`
	AvailablePlaybackActions struct {
		CanSkip      bool `json:"canSkip"`
		CanSkipBack  bool `json:"canSkipBack"`
		CanSeek      bool `json:"canSeek"`
		CanRepeat    bool `json:"canRepeat"`
		CanRepeatOne bool `json:"canRepeatOne"`
		CanCrossfade bool `json:"canCrossfade"`
		CanShuffle   bool `json:"canShuffle"`
	} `json:"availablePlaybackActions"`
}

type PlaybackMetadataResponse struct {
	Container   Container   `json:"container"`
	CurrentItem CurrentItem `json:"currentItem"`
	NextItem    NextItem    `json:"nextItem"`
	StreamInfo  string      `json:"streamInfo"`
}

// GetPlaybackStatus gets playback status
func (clt *Client) PlaybackGetStatus(id string) (*PlaybackStatusResponse, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback")

	body, err := clt.doApiRequest(http.MethodGet, url, nil)
	var resp PlaybackStatusResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Error("<client>Can't unmarshalling Favorites response : ", err)
		return nil, err
	}
	return &resp, nil
}

// PlaybackSet sends request to Sonos to play, pause or skip. Id is group id
func (clt *Client) PlaybackSet(val string, id string) (bool, error) {
	// change toggle_play_pause to togglePlayPause, next_track to skipToNextTrack and previous_track to skipToPreviousTrack
	if val == "toggle_play_pause" {
		val = "togglePlayPause"
	} else if val == "next_track" {
		val = "skipToNextTrack"
	} else if val == "previous_track" {
		val = "skipToPreviousTrack"
	}

	url := fmt.Sprintf("%s%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback/", val)

	binResponse, err := clt.doApiRequest(http.MethodPost, url, nil)
	var response interface{}
	err = json.Unmarshal(binResponse, &response)
	if err != nil {
		log.Error("Error when unmarshalling body: ", err)
		return false, err
	}
	// TODO : Check response
	return true, nil
}

// ModeSet sets new play mode
func (clt *Client) PlaybackModeSet(val map[string]bool, id string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback/playMode")
	// if one of the modes is repeat_one, change to repeatOne

	if _, ok := val["repeat_one"]; ok {
		val["repeatOne"] = val["repeat_one"]
		delete(val, "repeat_one")
	}

	mode := map[string]interface{}{
		"playModes": val,
	}

	binResponse, err := clt.doApiRequest(http.MethodPost, url, mode)
	var response interface{}
	err = json.Unmarshal(binResponse, &response)
	if err != nil {
		log.Error("Error when unmarshalling body: ", err)
		return false, err
	}
	// TODO : Check response
	return true, nil

}

func (clt *Client) GetMetadata(groupID string) (*PlaybackMetadataResponse, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/groups/", groupID, "/playbackMetadata")
	body, err := clt.doApiRequest(http.MethodGet, url, nil)
	var resp PlaybackMetadataResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Error("<client>Can't unmarshalling Favorites response : ", err)
		return nil, err
	}
	return &resp, nil
}
