package swift

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameImage struct {
	client *Client
}

func NewGameImage(client *Client) *GameImage {
	return &GameImage{
		client: client,
	}
}

func (gi *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, image *domain.GameImage) error {
	imageKey := gi.imageKey(image)

	var contentType string
	switch image.GetType() {
	case values.GameImageTypeJpeg:
		contentType = "image/jpeg"
	case values.GameImageTypePng:
		contentType = "image/png"
	case values.GameImageTypeGif:
		contentType = "image/gif"
	default:
		return fmt.Errorf("unsupported image type: %d", image.GetType())
	}

	err := gi.client.saveFile(
		ctx,
		imageKey,
		contentType,
		"",
		reader,
	)
	if errors.Is(err, ErrAlreadyExists) {
		return storage.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

// imageKey 変更時にはオブジェクトストレージのキーを変更する必要があるので要注意
func (gi *GameImage) imageKey(image *domain.GameImage) string {
	return fmt.Sprintf("images/%s", uuid.UUID(image.GetID()).String())
}
