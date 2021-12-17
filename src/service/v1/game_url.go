package v1

import "github.com/traPtitech/trap-collection-server/src/repository"

type GameURL struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
	gameURLRepository     repository.GameURL
}

func NewGameURL(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
	gameURLRepository repository.GameURL,
) *GameURL {
	return &GameURL{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
		gameURLRepository:     gameURLRepository,
	}
}
