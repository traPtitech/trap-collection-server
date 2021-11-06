package v1

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameVideo struct {
	db                  repository.DB
	gameRepository      repository.Game
	gameVideoRepository repository.GameVideo
	gameVideoStorage    storage.GameVideo
}

func NewGameVideo(
	db repository.DB,
	gameRepository repository.Game,
	gameVideoRepository repository.GameVideo,
	gameVideoStorage storage.GameVideo,
) *GameVideo {
	return &GameVideo{
		db:                  db,
		gameRepository:      gameRepository,
		gameVideoRepository: gameVideoRepository,
		gameVideoStorage:    gameVideoStorage,
	}
}
