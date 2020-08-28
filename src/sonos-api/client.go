package sonos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

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
		oauth2Client *edgeapp.FhOAuth2Client
		httpClient   *http.Client
		accessToken  string
		refreshToken string
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

	Favorite struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ImageURL    string `json:"imageUrl"`
		Service     struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		} `json:"service"`
	}

	Playlist struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		TrackCount int    `json:"trackCount"`
	}
)

func NewClient(env,accessToken,refreshToken string) *Client {
	authClient := edgeapp.NewFhOAuth2Client(sonosPartnerCode, sonosPartnerCode, env)

	return &Client{
		refreshToken: refreshToken,
		accessToken:  accessToken,
		oauth2Client: authClient,
		httpClient:   &http.Client{Timeout:30*time.Second}, // Very important to set timeout
	}
}

func (clt *Client) SetHubAuthToken(token string) {
	clt.oauth2Client.SetHubToken(token)
}

func (clt *Client) SetTokens(accessToken,refreshToken string) {
	clt.accessToken = accessToken
	clt.refreshToken = refreshToken
}

func (clt *Client) UpdateAuthParameters(mqttBrokerUri string) {
	clt.oauth2Client.SetParameters(mqttBrokerUri, "", "", 0, 0, 0, 0)
}

func (clt *Client) RefreshAccessToken(refreshToken string) (string, error) {
	if refreshToken == "" {
		refreshToken = clt.refreshToken
	}
	if refreshToken == "" {
		log.Error("<client> Empty refresh token")
		return "",fmt.Errorf("empty refresh token")
	}
	resp, err := clt.oauth2Client.ExchangeRefreshToken(refreshToken)
	if err != nil {
		log.Error("can't fetch new access token", err)
		return "", err
	}
	log.Debug("<client> New access token : ",resp.AccessToken)
	clt.accessToken = resp.AccessToken
	return resp.AccessToken, nil
}

func (clt *Client) GetHousehold() ([]Household, error) {
	url := fmt.Sprintf("%s%s", controlURL, "/v1/households")

	body,err := clt.doApiRequest(http.MethodGet,url,nil)
	if err != nil {
		return nil, err
	}
	type HouseholdsResponse struct {
		Households []Household `json:"households"`
	}
	var response HouseholdsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Error("Error when unmarshaling body: ", err)
		return nil, err
	}
	return response.Households, nil
}

// do a generic HTTP request
func (clt *Client) doHttpRequest(req *http.Request) ([]byte, error) {
	var err error
	var resp *http.Response
	for i:=0;i<3;i++ {
		resp, err = clt.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			break
		}else if resp.StatusCode == 401 {
			log.Info("Invalid token . Retrying")
			_,err= clt.RefreshAccessToken(clt.refreshToken)
			if err != nil {
				time.Sleep(time.Second*5)
			}else {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", clt.accessToken))
			}
		}else if resp.StatusCode != 200 {
			log.Error("Bad HTTP return code ", resp.StatusCode)
			return nil, err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Error("Bad HTTP return code ", resp.StatusCode)
		return nil, fmt.Errorf("bad HTTP return code %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func (clt *Client) doApiRequest(method,url string ,jsonRequest interface{})([]byte, error) {
	var requestBody io.Reader
	var err error
	if jsonRequest != nil {
		bytesRepresentation, err := json.Marshal(jsonRequest)
		if err != nil {
			log.Error("<client> Request marshalling error. Err:",err.Error())
			return nil, err
		}
		requestBody = bytes.NewBuffer(bytesRepresentation)
	}
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		log.Error(errors.Wrap(err, "getting metadata"))
		return nil, err
	}
	req.Header.Set("Content-Type","application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", clt.accessToken))
	return clt.doHttpRequest(req)
}

func (clt *Client) SetCorrectValue(val string) string {
	log.Debug(val)
	if val == "PLAYBACK_STATE_PLAYING" {
		return "play"
	} else if val == "PLAYBACK_STATE_PAUSED" {
		return "pause"
	}
	return "unknown"
}
