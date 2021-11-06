package v1

import "github.com/traPtitech/trap-collection-server/src/repository"

type GameVersion struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
}

func NewGameVersion(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
) *GameVersion {
	return &GameVersion{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
	}
}
