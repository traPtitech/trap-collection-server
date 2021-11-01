package v1

import "github.com/traPtitech/trap-collection-server/src/service"

type GameRole struct {
	session         *Session
	gameAuthService service.GameAuth
}

func NewGameRole(session *Session, gameAuthService service.GameAuth) *GameRole {
	return &GameRole{
		session:         session,
		gameAuthService: gameAuthService,
	}
}
