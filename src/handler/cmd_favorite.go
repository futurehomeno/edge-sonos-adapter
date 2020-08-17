package handler

import (
	"github.com/thingsplex/sonos/model"
)

type Favorite struct {
	states *model.States
}

func (fav *Favorite) FavoriteSet(val string) {
	return
}
