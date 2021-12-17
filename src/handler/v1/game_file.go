package v1

import "github.com/traPtitech/trap-collection-server/src/service"

type GameFile struct {
	gameFileService service.GameFile
}

func NewGameFile(gameFileService service.GameFile) *GameFile {
	return &GameFile{
		gameFileService: gameFileService,
	}
}
