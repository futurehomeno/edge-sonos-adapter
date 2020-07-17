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

// Volume attributes
type Volume struct {
	states *model.States
}

// VolumeSet sends request to Sonos to set new volume lvl
func (vol *Volume) VolumeSet(val int64, id string, accessToken string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/groupVolume")

	volume := map[string]interface{}{
		"volume": val,
	}
	bytesRepresentation, err := json.Marshal(volume)
	if err != nil {
		log.Error("bytesRepresentation error on VolumeSet: ", err)
		return false, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Error("Error when setting volume: ", err)
		return false, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on VolumeSet: ", err)
		return false, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on VolumeSet: ", err)
		return false, err
	}
	err = json.Unmarshal(body, &vol)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return false, err
	}
	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return false, err
	}

	return true, nil
}
