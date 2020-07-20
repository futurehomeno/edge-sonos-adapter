package handler

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/sonos/sonos-api"
)

type Id struct {
}

// FindGroupFromPlayer finds the GroupID of the Group that contains the given Player
func (id *Id) FindGroupFromPlayer(playerID string, groups []sonos.Group) (string, error) {
	// I have player ID from parameter.
	// Iterate over all playersID's in all groups until I find a match
	standardID := "RINCON_"
	ID := fmt.Sprintf("%s%s", standardID, playerID)
	for i := 0; i < len(groups); i++ {
		log.Debug("i")
		for p := 0; p < len(groups[i].PlayersIds); p++ {
			log.Debug("p")
			if ID == groups[i].PlayersIds[p] {
				log.Debug("found")
				return groups[i].GroupId, nil
			}
		}
	}
	err := fmt.Errorf("Could not find a group which contains player")
	return "", err
}