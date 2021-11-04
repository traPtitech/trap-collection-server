package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameImage struct {
	imageRootPath    string
	directoryManager *DirectoryManager
}

func NewGameImage(directoryManager *DirectoryManager) (*GameImage, error) {
	imageRootPath, err := directoryManager.setupDirectory("images")
	if err != nil {
		return nil, fmt.Errorf("failed to setup directory: %w", err)
	}

	return &GameImage{
		imageRootPath:    imageRootPath,
		directoryManager: directoryManager,
	}, nil
}

func (gi *GameImage) SaveGameImage(ctx context.Context, reader io.Reader, image *domain.GameImage) error {
	imagePath := path.Join(gi.imageRootPath, uuid.UUID(image.GetID()).String())

	_, err := os.Stat(imagePath)
	if err == nil {
		return storage.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	f, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}
