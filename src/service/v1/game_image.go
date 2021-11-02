package v1

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameImage struct {
	db                  repository.DB
	gameRepository      repository.Game
	gameImageRepository repository.GameImage
	gameImageStorage    storage.GameImage
}

func NewGameImage(
	db repository.DB,
	gameRepository repository.Game,
	gameImageRepository repository.GameImage,
	gameImageStorage storage.GameImage,
) *GameImage {
	return &GameImage{
		db:                  db,
		gameRepository:      gameRepository,
		gameImageRepository: gameImageRepository,
		gameImageStorage:    gameImageStorage,
	}
}
