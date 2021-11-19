package v1

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameFile struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
	gameFileRepository    repository.GameFile
	gameFileStorage       storage.GameFile
}

func NewGameFile(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
	gameFileRepository repository.GameFile,
	gameFileStorage storage.GameFile,
) *GameFile {
	return &GameFile{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
		gameFileRepository:    gameFileRepository,
		gameFileStorage:       gameFileStorage,
	}
}
