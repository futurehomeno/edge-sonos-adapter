package sonos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type AudioClipRequest struct {
	Name      string `json:"name"`
	AppID     string `json:"appId"`
	StreamURL string `json:"streamUrl"`
	ClipType  string `json:"clipType"`
	Volume    int    `json:"volume"`
	Priority  string  `json:"priority"`
}

//AudioClipLoad - Plays audio clip from URL and plays it with set volume. URLs can be local and global.Supported audio format : mp3,wav,etc.
func (clt *Client) AudioClipLoad(clipUrl string,volume int, playerId string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/players/", playerId, "/audioClip")

	body := AudioClipRequest{
		Name:      "INTRUSION_ALARM_MSG",
		AppID:     "futurehome.edge.app",
		StreamURL: clipUrl,
		ClipType:  "CUSTOM",
		Volume:    volume,
		Priority: "HIGH",
	}

	type responseT struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		AppId    string `json:"appId"`
		Priority string `json:"priority"`
		ClipType string `json:"clipType"`
	}
	var response responseT

	binResponse, err := clt.doApiRequest(http.MethodPost, url, body)
	err = json.Unmarshal(binResponse, &response)
	if err != nil {
		log.Error("Error when unmarshalling body: ", err)
		return false, err
	}
	return true, err
}
