package sonos

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)
type GroupsAndPlayersResponse struct {
	Players []Player `json:"players"`
	Groups  []Group `json:"groups"`
}

func (clt *Client) GetGroupsAndPlayers(HouseholdID string) ([]Group, []Player, error) {
	url := fmt.Sprintf("%s%s%s%s", controlURL, "/v1/households/", HouseholdID, "/groups")

	body , err:= clt.doApiRequest(http.MethodGet,url,nil)
	var resp GroupsAndPlayersResponse

	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Error("Can't unmarshal GroupsAndPlayers body: ", err)
		return nil, nil, err
	}
	for i := 0; i < len(resp.Groups); i++ {
		IDs := strings.Split(resp.Groups[i].GroupId, "_")[1]
		resp.Groups[i].FimpId = strings.Split(IDs, ":")[0]
		resp.Groups[i].OnlyGroupId = strings.Split(IDs, ":")[1]
	}
	for i := 0; i < len(resp.Players); i++ {
		resp.Players[i].FimpId = strings.Split(resp.Players[i].Id, "_")[1]
	}

	return resp.Groups, resp.Players, nil
}

// FindGroupFromPlayer finds the GroupID of the Group that contains the given Player
func (clt *Client) FindGroupFromPlayer(playerID string, groups []Group) (string, error) {
	// I have player ID from parameter.
	// Iterate over all playersID's in all groups until I find a match
	standardID := "RINCON_"
	ID := fmt.Sprintf("%s%s", standardID, playerID)
	for i := 0; i < len(groups); i++ {
		for p := 0; p < len(groups[i].PlayersIds); p++ {
			if ID == groups[i].PlayersIds[p] {
				return groups[i].GroupId, nil
			}
		}
	}
	err := fmt.Errorf("could not find a group which contains player")
	return "", err
}