package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// Volume attributes
type Volume struct {
	status string
}

// VolumeSet sends request to Sonos to set new volume lvl
func (vol *Volume) VolumeSet(val int64, id string, accessToken string) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/groupVolume")

	volume := map[string]interface{}{
		"volume": val,
	}
	bytesRepresentation, err := json.Marshal(volume)
	if err != nil {
		log.Error("bytesRepresentation error on VolumeSet: ", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Error("Error when setting volume: ", err)
		return
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on VolumeSet: ", err)
	}
	log.Debug("volumeSet first resp: ", resp)
}
