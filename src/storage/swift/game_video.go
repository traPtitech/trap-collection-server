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

type GameVideo struct {
	client *Client
}

func NewGameVideo(client *Client) *GameVideo {
	return &GameVideo{
		client: client,
	}
}

func (gv *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, video *domain.GameVideo) error {
	videoKey := gv.videoKey(video)

	var contentType string
	switch video.GetType() {
	case values.GameVideoTypeMp4:
		contentType = "video/mp4"
	default:
		return fmt.Errorf("unsupported video type: %d", video.GetType())
	}

	err := gv.client.saveFile(
		ctx,
		videoKey,
		contentType,
		"",
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

func (gv *GameVideo) GetGameVideo(ctx context.Context, writer io.Writer, video *domain.GameVideo) error {
	videoKey := gv.videoKey(video)

	err := gv.client.loadFile(
		ctx,
		videoKey,
		writer,
	)
	if errors.Is(err, ErrNotFound) {
		return storage.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	return nil
}

func (gv *GameVideo) videoKey(video *domain.GameVideo) string {
	return fmt.Sprintf("videos/%s", uuid.UUID(video.GetID()).String())
}
