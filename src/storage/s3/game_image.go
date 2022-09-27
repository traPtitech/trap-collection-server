package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

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

func (gi *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, imageID values.GameImageID) error {
	imageKey := gi.imageKey(imageID)

	err := gi.client.saveFile(
		ctx,
		imageKey,
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

func (gi *GameImage) GetTempURL(ctx context.Context, image *domain.GameImage, expires time.Duration) (values.GameImageTmpURL, error) {
	filekey := gi.imageKey(image.GetID())

	url, err := gi.client.createTempURL(ctx, filekey, expires)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get temp url: %w", err)
	}

	return url, nil
}

// imageKey 変更時にはオブジェクトストレージのキーを変更する必要があるので要注意
func (gi *GameImage) imageKey(imageID values.GameImageID) string {
	return fmt.Sprintf("images/%s", uuid.UUID(imageID).String())
}
