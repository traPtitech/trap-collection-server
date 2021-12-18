package v1

import "github.com/traPtitech/trap-collection-server/src/service"

type GameURL struct {
	gameURLService service.GameURL
}

func NewGameURL(gameURLService service.GameURL) *GameURL {
	return &GameURL{
		gameURLService: gameURLService,
	}
}
