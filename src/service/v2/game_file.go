package v2

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameFile struct {
	db                 repository.DB
	gameRepository     repository.GameV2
	gameFileRepository repository.GameFileV2
	gameFileStorage    storage.GameFile
}

func NewGameFile(
	db repository.DB,
	gameRepository repository.GameV2,
	gameFileRepository repository.GameFileV2,
	gameFileStorage storage.GameFile,
) *GameFile {
	return &GameFile{
		db:                 db,
		gameRepository:     gameRepository,
		gameFileRepository: gameFileRepository,
		gameFileStorage:    gameFileStorage,
	}
}
