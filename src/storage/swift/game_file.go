package swift

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
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

func (gf *GameFile) SaveGameFile(ctx context.Context, reader io.Reader, file *domain.GameFile) error {
	fileKey := gf.fileKey(file)

	contentType := "application/zip"

	err := gf.client.saveFile(
		ctx,
		fileKey,
		contentType,
		hex.EncodeToString([]byte(file.GetHash())),
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

func (gf *GameFile) GetGameFile(ctx context.Context, writer io.Writer, file *domain.GameFile) error {
	fileKey := gf.fileKey(file)

	err := gf.client.loadFile(
		ctx,
		fileKey,
		writer,
	)
	if errors.Is(err, ErrNotFound) {
		return storage.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	return nil
}

func (gf *GameFile) fileKey(file *domain.GameFile) string {
	return fmt.Sprintf("files/%s", uuid.UUID(file.GetID()).String())
}
