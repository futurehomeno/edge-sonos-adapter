package sonos

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	controlURL = "https://api.ws.sonos.com/control/api"
)

type Config struct {
	ErrorCode  int    `json:"errorCode"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Success    bool   `json:"success"`
}

type Client struct {
	httResponse *http.Response
	Households  []Household `json:"households"`
	Groups      []Group     `json:"groups"`
	Players     []Player    `json:"players"`
}

type Household struct {
	ID string `json:"id"`
}

type Group struct {
	GroupId       string        `json:"id"`
	Name          string        `json:"name"`
	CoordinatorId string        `json:"coordinatorId"`
	PlaybackState string        `json:"playbackState"`
	PlayersIds    []interface{} `json:"playerIds"`
}

type Player struct {
	Id            string        `json:"id"`
	Name          string        `json:"name"`
	Icon          string        `json:"icon"`
	SWVersion     string        `json:"softwareVersion"`
	DeviceIds     []interface{} `json:"deviceIds"`
	APIVersion    string        `json:"apiVersion"`
	MinAPIVersion string        `json:"minApiVersion"`
	Capabilities  []interface{} `json:"capabilities"`
}

func (c *Client) GetHousehold(accessToken string) (*Client, error) {
	url := fmt.Sprintf("%s%s", controlURL, "v1/households")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(fmt.Errorf("Can't get households, error: ", err))
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	resp, err := http.DefaultClient.Do(req)
	processHTTPResponse(resp, err, c)

	return c, nil
}

func (c *Client) GetPlayersAndGroups(accessToken string, HouseholdID string) (*Client, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "v1/households", HouseholdID, "groups")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(fmt.Errorf("Can't get players and groups, error: ", err))
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	resp, err := http.DefaultClient.Do(req)
	log.Debug(resp)
	return nil, nil
}

func processHTTPResponse(resp *http.Response, err error, holder interface{}) error {
	if err != nil {
		log.Error(fmt.Errorf("API does not respond"))
		return err
	}
	defer resp.Body.Close()
	// check http return code
	if resp.StatusCode != 200 {
		//bytes, _ := ioutil.ReadAll(resp.Body)
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return fmt.Errorf("Bad HTTP return code %d", resp.StatusCode)
	}

	// Unmarshall response into given struct
	if err = json.NewDecoder(resp.Body).Decode(holder); err != nil {
		return err
	}
	return nil
}
