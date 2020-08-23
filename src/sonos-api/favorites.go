package sonos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (clt *Client) FavoritesGet(HouseholdID string) ([]Favorite, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/households/", HouseholdID, "/favorites")

	type FavoritesResponse struct {
		Version   string `json:"version"`
		Favorites []Favorite `json:"items"`
	}

	body , err:= clt.doApiRequest(http.MethodGet,url,nil)
	var resp FavoritesResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Error("<client>Can't unmarshalling Favorites response : ", err)
		return nil, err
	}
	return resp.Favorites, nil

}

func (clt *Client) FavoriteSet(val string, id string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/favorites")

	body := map[string]interface{}{
		"action":           "INSERT_NEXT",
		"favoriteId":       val,
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
