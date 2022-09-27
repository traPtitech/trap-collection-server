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

type GameVideo struct {
	client *Client
}

func NewGameVideo(client *Client) *GameVideo {
	return &GameVideo{
		client: client,
	}
}

func (gv *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, videoID values.GameVideoID) error {
	videoKey := gv.videoKey(videoID)

	err := gv.client.saveFile(
		ctx,
		videoKey,
		reader,
	)
	if errors.Is(err, ErrAlreadyExists) {
		return storage.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("failed to save video: %w", err)
	}

	return nil
}

func (gv *GameVideo) GetTempURL(ctx context.Context, video *domain.GameVideo, expires time.Duration) (values.GameVideoTmpURL, error) {
	fileKey := gv.videoKey(video.GetID())

	url, err := gv.client.createTempURL(ctx, fileKey, expires)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get temp url: %w", err)
	}

	return url, nil
}

func (gv *GameVideo) videoKey(videoID values.GameVideoID) string {
	return fmt.Sprintf("videos/%s", uuid.UUID(videoID).String())
}
