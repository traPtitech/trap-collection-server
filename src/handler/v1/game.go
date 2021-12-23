package v1

import "github.com/traPtitech/trap-collection-server/src/service"

type Game struct {
	session     *Session
	gameService service.Game
}

func NewGame(
	session *Session,
	gameService service.Game,
) *Game {
	return &Game{
		session:     session,
		gameService: gameService,
	}
}
