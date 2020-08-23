package sonos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (clt *Client) PlaylistsGet( HouseholdID string) ([]Playlist, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/households/", HouseholdID, "/playlists")

	type PlaylistResponse struct {
		Playlists []Playlist `json:"playlists"`
	}

	body , err:= clt.doApiRequest(http.MethodGet,url,nil)
	var resp PlaylistResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}
	return resp.Playlists, nil
}

func (clt *Client) PlaylistSet(val string, id string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playlists")

	body := map[string]interface{}{
		"action":           "REPLACE",
		"playlistId":       val,
		"playOnCompletion": true,
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