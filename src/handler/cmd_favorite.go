package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/sonos/model"
)

type Favorite struct {
	states *model.States
}

func (fav *Favorite) FavoriteSet(val string, id string, accessToken string) (bool, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/favorites")

	body := map[string]interface{}{
		"action":           "INSERT_NEXT",
		"favoriteId":       val,
		"playOnCompletion": true,
	}

	bytesRepresentation, err := json.Marshal(body)
	if err != nil {
		log.Error(err)
		return false, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Error("Error when setting favorite: ", err)
		return false, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on FavoriteSet: ", err)
		return false, err
	}

	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return false, fmt.Errorf("%s%s", "Bad HTTP return code ", strconv.Itoa(resp.StatusCode))
	}

	return true, nil
}
