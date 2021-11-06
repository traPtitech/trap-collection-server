package swift

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
)

type GameVideo struct {
	client *Client
}

func NewGameVideo(client *Client) *GameVideo {
	return &GameVideo{
		client: client,
	}
}

func (gv *GameVideo) videoKey(video *domain.GameVideo) string {
	return fmt.Sprintf("videos/%s", uuid.UUID(video.GetID()).String())
}
