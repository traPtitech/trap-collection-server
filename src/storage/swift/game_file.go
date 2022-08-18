package swift

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

type GameFile struct {
	client *Client
}

func NewGameFile(client *Client) *GameFile {
	return &GameFile{
		client: client,
	}
}

func (gf *GameFile) SaveGameFile(ctx context.Context, reader io.Reader, fileID values.GameFileID) error {
	fileKey := gf.fileKey(fileID)

	contentType := "application/zip"

	err := gf.client.saveFile(
		ctx,
		fileKey,
		contentType,
		"",
		reader,
	)
	if errors.Is(err, ErrAlreadyExists) {
		return storage.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

func (gf *GameFile) GetTempURL(ctx context.Context, file *domain.GameFile, expires time.Duration) (values.GameFileTmpURL, error) {
	fileKey := gf.fileKey(file.GetID())

	url, err := gf.client.createTempURL(ctx, fileKey, expires)
	if errors.Is(err, ErrNotFound) {
		return nil, storage.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get temp url: %w", err)
	}

	return url, nil
}

func (gf *GameFile) fileKey(fileID values.GameFileID) string {
	return fmt.Sprintf("files/%s", uuid.UUID(fileID).String())
}
