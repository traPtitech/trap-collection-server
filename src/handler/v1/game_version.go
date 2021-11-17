package v1

import (
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVersion struct {
	gameVersionService service.GameVersion
}

func NewGameVersion(gameVersionService service.GameVersion) *GameVersion {
	return &GameVersion{
		gameVersionService: gameVersionService,
	}
}
