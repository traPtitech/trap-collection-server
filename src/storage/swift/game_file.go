package swift

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
)

type GameFile struct {
	client *Client
}

func NewGameFile(client *Client) *GameFile {
	return &GameFile{
		client: client,
	}
}

func (gf *GameFile) fileKey(file *domain.GameFile) string {
	return fmt.Sprintf("files/%s", uuid.UUID(file.GetID()).String())
}
