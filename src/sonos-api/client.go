package sonos

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/futurehomeno/fimpgo/edgeapp"
	log "github.com/sirupsen/logrus"
)

const (
	controlURL       = "https://api.ws.sonos.com/control/api"
	sonosPartnerCode = "sonos"
)

type (
	Config struct {
		ErrorCode  int    `json:"errorCode"`
		Message    string `json:"message"`
		StatusCode int    `json:"statusCode"`
		Success    bool   `json:"success"`
	}

	Client struct {
		futurehomeOauth2Client *edgeapp.FhOAuth2Client

		Households  []Household `json:"households"`
		Groups      []Group     `json:"groups"`
		Players     []Player    `json:"players"`
		Container   Container   `json:"container"`
		CurrentItem CurrentItem `json:"currentItem"`
		NextItem    NextItem    `json:"nextItem"`
		StreamInfo  string      `json:"streamInfo"`

		PlaybackState string `json:"playbackState"`

		Volume int  `json:"volume"`
		Muted  bool `json:"muted"`
		Fixed  bool `json:"fixed"`

		Version   string     `json:"version"`
		Favorites []Favorite `json:"favorites"`

		Playlists []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Type       string `json:"type"`
			TrackCount int    `json:"trackCount"`
		} `json:"playlists"`
		PlayModes struct {
			Repeat    bool `json:"repeat"`
			RepeatOne bool `json:"repeatOne"`
			Shuffle   bool `json:"shuffle"`
			Crossfade bool `json:"crossfade"`
		}
	}

	Household struct {
		ID string `json:"id"`
	}

	Group struct {
		GroupId       string `json:"id"`
		OnlyGroupId   string
		FimpId        string
		Name          string        `json:"name"`
		CoordinatorId string        `json:"coordinatorId"`
		PlaybackState string        `json:"playbackState"`
		PlayersIds    []interface{} `json:"playerIds"`
	}

	Player struct {
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

	Container struct {
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

	CurrentItem struct {
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

	NextItem struct {
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
)

func NewClient(env string) *Client {
	return &Client{
		futurehomeOauth2Client: edgeapp.NewFhOAuth2Client(sonosPartnerCode, sonosPartnerCode, env),
	}
}

type Favorite struct {
	Items []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Service     struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"service"`
	} `json:"items"`
}

type Playlist struct {
	Playlists []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		TrackCount int    `json:"trackCount"`
	} `json:"playlists"`
}

func (c *Client) RefreshAccessToken(refreshToken string, mqttServerUri string) (string, error) {
	c.futurehomeOauth2Client.SetParameters(mqttServerUri, "", "", 0, 0, 0, 0)
	err := c.futurehomeOauth2Client.Init()
	if err != nil {
		log.Error("failed to init sync client")
		return "", err
	}

	log.Debug(refreshToken)

	resp, err := c.futurehomeOauth2Client.ExchangeRefreshToken(refreshToken)
	if err != nil {
		log.Error("can't fetch new access token", err)
		return "", err
	}
	return resp.AccessToken, nil
}

func (c *Client) GetHousehold(accessToken string) ([]Household, error) {
	url := fmt.Sprintf("%s%s", controlURL, "/v1/households")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(errors.Wrap(err, "can't get households, error: "))
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
	defer closeBody(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(errors.Wrap(err, "reading body"))
		return nil, err
	}
	err = json.Unmarshal(body, &c)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}

	return c.Households, nil
}

func (fv *Favorite) GetFavorites(accessToken string, HouseholdID string) (*Favorite, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/households/", HouseholdID, "/favorites")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(fmt.Errorf("Can't get favorites, error: ", err))
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetFavorites: ", err)
		return nil, err
	}
	log.Debug("resp: ", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetFavorites: ", err)
		return nil, err
	}
	log.Debug("body: ", body)
	err = json.Unmarshal(body, &fv)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}
	return fv, nil
}

func (c *Client) GetPlaylists(accessToken string, HouseholdID string) (*Client, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/households/", HouseholdID, "/playlists")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(fmt.Errorf("Can't get playlists, error: ", err))
		return nil, err
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", os.ExpandEnv(fmt.Sprintf("%s%s", "Bearer ", accessToken)))
	log.Debug("New Request: ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error when DefaultClient.Do on GetPlaylists: ", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error when ioutil.ReadAll on GetPlaylists: ", err)
		return nil, err
	}
	err = json.Unmarshal(body, &c.Playlists)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}
	return c, nil
}

func (c *Client) GetGroupsAndPlayers(accessToken string, HouseholdID string) ([]Group, []Player, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/households/", HouseholdID, "/groups")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error(errors.Wrap(err, "getting players and groups, error: "))
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
	defer closeBody(resp.Body)

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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error(errors.Wrap(err, "getting metadata"))
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
	defer closeBody(resp.Body)

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
	defer closeBody(resp.Body)

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
	defer closeBody(resp.Body)

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

func closeBody(c io.Closer) {
	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		log.Error(errors.Wrap(err, "closing body"))
	}
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
		return fmt.Errorf("bad HTTP return code %d", resp.StatusCode)
	}

	// Unmarshall response into given struct
	if err = json.NewDecoder(resp.Body).Decode(holder); err != nil {
		return err
	}
	return nil
}

func (c *Client) SetCorrectValue(val string) string {
	log.Debug(val)
	if val == "PLAYBACK_STATE_PLAYING" {
		return "play"
	} else if val == "PLAYBACK_STATE_PAUSED" {
		return "pause"
	}
	return "unknown"
}
