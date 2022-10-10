package v2

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

var _ service.GameVersionV2 = &GameVersion{}

type GameVersion struct {
	db repository.DB
	// TODO: v2のgameRepositoryに変更
	gameRepository        repository.Game
	gameImageRepository   repository.GameImageV2
	gameVideoRepository   repository.GameVideoV2
	gameFileRepository    repository.GameFileV2
	gameVersionRepository repository.GameVersionV2
}

func NewGameVersion(
	db repository.DB,
	gameRepository repository.Game,
	gameImageRepository repository.GameImageV2,
	gameVideoRepository repository.GameVideoV2,
	gameFileRepository repository.GameFileV2,
	gameVersionRepository repository.GameVersionV2,
) *GameVersion {
	return &GameVersion{
		db:                    db,
		gameRepository:        gameRepository,
		gameImageRepository:   gameImageRepository,
		gameVideoRepository:   gameVideoRepository,
		gameFileRepository:    gameFileRepository,
		gameVersionRepository: gameVersionRepository,
	}
}
