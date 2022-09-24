package v2

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

var _ service.GameImageV2 = &GameImage{}

type GameImage struct {
	db                  repository.DB
	gameRepository      repository.Game
	gameImageRepository repository.GameImageV2
	gameImageStorage    storage.GameImage
}

func NewGameImage(
	db repository.DB,
	gameRepository repository.Game,
	gameImageRepository repository.GameImageV2,
	gameImageStorage storage.GameImage,
) *GameImage {
	return &GameImage{
		db:                  db,
		gameRepository:      gameRepository,
		gameImageRepository: gameImageRepository,
		gameImageStorage:    gameImageStorage,
	}
}
