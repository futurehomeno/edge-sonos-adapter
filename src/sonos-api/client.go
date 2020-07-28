package sonos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/futurehomeno/fimpgo/edgeapp"
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
	Container   Container   `json:"container"`
	CurrentItem CurrentItem `json:"currentItem"`
	NextItem    NextItem    `json:"nextItem"`
	StreamInfo  string      `json:"streamInfo"`

	PlaybackState string `json:"playbackState"`

	PlayModes struct {
		Repeat    bool `json:"repeat"`
		RepeatOne bool `json:"repeatOne"`
		Shuffle   bool `json:"shuffle"`
		Crossfade bool `json:"crossfade"`
	}

	Volume int  `json:"volume"`
	Muted  bool `json:"muted"`
	Fixed  bool `json:"fixed"`
}

type Household struct {
	ID string `json:"id"`
}

type Group struct {
	GroupId       string `json:"id"`
	OnlyGroupId   string
	FimpId        string
	Name          string        `json:"name"`
	CoordinatorId string        `json:"coordinatorId"`
	PlaybackState string        `json:"playbackState"`
	PlayersIds    []interface{} `json:"playerIds"`
}

type Player struct {
	Id             string `json:"id"`
	FimpId         string
	Name           string        `json:"name"`
	WebSocketUrl   string        `json:"websocketUrl"`
	SWVersion      string        `json:"softwareVersion"`
	APIVersion     string        `json:"apiVersion"`
	MinAPIVersion  string        `json:"minApiVersion"`
	IsUnregistered bool          `json:"isUnregistered"`
	Capabilities   []interface{} `json:"capabilities"`
	DeviceIds      []interface{} `json:"deviceIds"`
	Icon           string        `json:"icon"`
}

type Container struct {
	Name string `json:"name"`
	Type string `json:"type"`
	ID   struct {
		ServiceID string `json:"serviceId"`
		ObjectID  string `json:"objectId"`
		AccountID string `json:"accountId"`
	} `json:"id"`
	Service struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"service"`
	ImageURL string `json:"imageUrl"`
}

type CurrentItem struct {
	Track struct {
		Type     string `json:"type"`
		Name     string `json:"name"`
		ImageURL string `json:"imageUrl"`
		Album    struct {
			Name string `json:"name"`
		} `json:"album"`
		Artist struct {
			Name string `json:"name"`
		} `json:"artist"`
		ID struct {
			ServiceID string `json:"serviceId"`
			ObjectID  string `json:"objectId"`
			AccountID string `json:"accountId"`
		} `json:"id"`
		Service struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"service"`
		DurationMillis int      `json:"durationMillis"`
		Tags           []string `json:"tags"`
	} `json:"track"`
}

type NextItem struct {
	Track struct {
		Type     string `json:"type"`
		Name     string `json:"name"`
		ImageURL string `json:"imageUrl"`
		Album    struct {
			Name string `json:"name"`
		} `json:"album"`
		Artist struct {
			Name string `json:"name"`
		} `json:"artist"`
		ID struct {
			ServiceID string `json:"serviceId"`
			ObjectID  string `json:"objectId"`
			AccountID string `json:"accountId"`
		} `json:"id"`
		Service struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"service"`
		DurationMillis int      `json:"durationMillis"`
		Tags           []string `json:"tags"`
	} `json:"track"`
}

func (c *Client) RefreshAccessToken(refreshToken string, mqttServerUri string, env string) (string, error) {
	client := edgeapp.NewFhOAuth2Client("sonos", "sonos", env)
	client.SetParameters(mqttServerUri, "", "", 0, 0, 0, 0)
	err := client.Init()
	if err != nil {
		log.Error("Failed to init sync client")
	}
	log.Debug(refreshToken)
	resp, err := client.ExchangeRefreshToken(refreshToken)
	if err != nil {
		log.Error("Can't fetch new access token", err)
		return "", err
	}
	return resp.AccessToken, nil
}

func (c *Client) GetHousehold(accessToken string) ([]Household, error) {
	url := fmt.Sprintf("%s%s", controlURL, "/v1/households")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(fmt.Errorf("Can't get households, error: ", err))
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetHousehold: ", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.Readall on GetHousehold: ", err)
		return nil, err
	}
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}

	return c.Households, nil
}

func (c *Client) GetGroupsAndPlayers(accessToken string, HouseholdID string) ([]Group, []Player, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/households/", HouseholdID, "/groups")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(fmt.Errorf("Can't get players and groups, error: ", err))
		return nil, nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetPlayersAndGroups: ", err)
		return nil, nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetGroupsAndPlayers: ", err)
		return nil, nil, err
	}
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, nil, err
	}
	for i := 0; i < len(c.Groups); i++ {
		IDs := strings.Split(c.Groups[i].GroupId, "_")[1]
		c.Groups[i].FimpId = strings.Split(IDs, ":")[0]
		c.Groups[i].OnlyGroupId = strings.Split(IDs, ":")[1]
	}
	for i := 0; i < len(c.Players); i++ {
		c.Players[i].FimpId = strings.Split(c.Players[i].Id, "_")[1]
	}

	return c.Groups, c.Players, nil
}

func (c *Client) GetMetadata(accessToken string, groupID string) (*Client, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/groups/", groupID, "/playbackMetadata")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(fmt.Errorf("Can't get metadata, error: ", err))
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetMetadata: ", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetGroupsAndPlayers: ", err)
		return nil, err
	}
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}

	return c, nil
}

// GetPlaybackStatus gets playback status
func (c *Client) GetPlaybackStatus(id string, accessToken string) (*Client, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/playback")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Error when getting playback status: ", err)
		return nil, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetPlaybackStatus: ", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetPlaybackStatus: ", err)
		return nil, err
	}
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return nil, fmt.Errorf("%s%s", "Bad HTTP return code ", strconv.Itoa(resp.StatusCode))
	}

	return c, nil
}

func (c *Client) GetVolume(id string, accessToken string) (*Client, error) {
	url := fmt.Sprintf("%s%s%s", "https://api.ws.sonos.com/control/api/v1/groups/", id, "/groupVolume")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Error when getting volume status: ", err)
		return nil, err
	}
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	req.Header.Set("Content-Type", "application/json")
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetVolume: ", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetVolume: ", err)
		return nil, err
	}
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return nil, fmt.Errorf("%s%s", "Bad HTTP return code ", strconv.Itoa(resp.StatusCode))
	}

	return c, nil
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
