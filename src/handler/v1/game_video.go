package v1

import "github.com/traPtitech/trap-collection-server/src/service"

type GameVideo struct {
	gameVideoService service.GameVideo
}

func NewGameVideo(gameVideoService service.GameVideo) *GameVideo {
	return &GameVideo{
		gameVideoService: gameVideoService,
	}
}
