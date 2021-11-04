package swift

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
)

type GameImage struct {
	client *Client
}

func NewGameImage(client *Client) *GameImage {
	return &GameImage{
		client: client,
	}
}

// imageKey 変更時にはオブジェクトストレージのキーを変更する必要があるので要注意
func (gi *GameImage) imageKey(image *domain.GameImage) string {
	return fmt.Sprintf("images/%s", uuid.UUID(image.GetID()).String())
}
