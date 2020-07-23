package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/sonos/model"
)

// Mute attributes
type Mute struct {
	states *model.States
}

// MuteSet sends request to Sonos to mute or unmute speaker
func (mute *Mute) MuteSet(val bool, id string, accessToken string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/groupVolume/mute")

	ume := map[string]interface{}{
		"muted": val,
	}
	bytesRepresentation, err := json.Marshal(ume)
	if err != nil {
		log.Error("bytesRepresentation error on MuteSet: ", err)
		return false, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Error("Error when setting ume: ", err)
		return false, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on MuteSet: ", err)
		return false, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on MuteSet: ", err)
		return false, err
	}
	err = json.Unmarshal(body, &mute)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return false, err
	}
	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return false, fmt.Errorf("%s%s", "Bad HTTP return code ", strconv.Itoa(resp.StatusCode))
	}

	return true, nil
}
