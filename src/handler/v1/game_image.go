package v1

import (
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameImage struct {
	gameImageService service.GameImage
}

func NewGameImage(gameImageService service.GameImage) *GameImage {
	return &GameImage{
		gameImageService: gameImageService,
	}
}
