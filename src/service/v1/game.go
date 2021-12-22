package v1

import "github.com/traPtitech/trap-collection-server/src/repository"

type Game struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
	userUtils             *UserUtils
}

func NewGame(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
	userUtils *UserUtils,
) *Game {
	return &Game{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
		userUtils:             userUtils,
	}
}
