package sonos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type VolumeResponse struct {
	Volume int  `json:"volume"`
	Muted  bool `json:"muted"`
	Fixed  bool `json:"fixed"`
}

func (clt *Client) VolumeGet(id string) (*VolumeResponse, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/groupVolume")
	body, err := clt.doApiRequest(http.MethodGet, url, nil)
	var resp VolumeResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Error("<client>Can't unmarshalling Volume response : ", err)
		return nil, err
	}
	return &resp, nil
}

// VolumeSet sends request to Sonos to set new volume lvl
func (clt *Client) VolumeSet(val int64, id string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/groupVolume")

	body := map[string]interface{}{
		"volume": val,
	}

	binResponse, err := clt.doApiRequest(http.MethodPost, url, body)
	var response map[string]string
	err = json.Unmarshal(binResponse, &response)
	if err != nil {
		log.Error("Error when unmarshalling body: ", err)
		return false, err
	}
	// TODO : Check response
	return true, nil
}

// MuteSet sends request to Sonos to mute or unmute speaker
func (clt *Client) VolumeMuteSet(val bool, id string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/groupVolume/mute")

	body := map[string]interface{}{
		"muted": val,
	}
	binResponse, err := clt.doApiRequest(http.MethodPost, url, body)
	var response map[string]string
	err = json.Unmarshal(binResponse, &response)
	if err != nil {
		log.Error("Error when unmarshalling body: ", err)
		return false, err
	}
	// TODO : Check response
	return true, nil
}

