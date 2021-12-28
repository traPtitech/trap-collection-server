package local

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type GameFile struct {
	fileRootPath     string
	directoryManager *DirectoryManager
}

func NewGameFile(directoryManager *DirectoryManager) (*GameFile, error) {
	fileRootPath, err := directoryManager.setupDirectory("files")
	if err != nil {
		return nil, fmt.Errorf("failed to setup directory: %w", err)
	}

	return &GameFile{
		fileRootPath:     fileRootPath,
		directoryManager: directoryManager,
	}, nil
}

func (gf *GameFile) SaveGameFile(ctx context.Context, reader io.Reader, fileID values.GameFileID) error {
	filePath := path.Join(gf.fileRootPath, uuid.UUID(fileID).String())

	_, err := os.Stat(filePath)
	if err == nil {
		return storage.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	f, err := os.Create(filePath)
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

func (gf *GameFile) GetGameFile(ctx context.Context, writer io.Writer, file *domain.GameFile) error {
	filePath := path.Join(gf.fileRootPath, uuid.UUID(file.GetID()).String())

	f, err := os.Open(filePath)
	if errors.Is(err, fs.ErrNotExist) {
		return storage.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(writer, f)
	if err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	return nil
}

func (gf *GameFile) GetTempURL(ctx context.Context, file *domain.GameFile, expires time.Duration) (values.GameFileTmpURL, error) {
	tmpURL, err := url.Parse(fmt.Sprintf("file://%s", path.Join(gf.fileRootPath, uuid.UUID(file.GetID()).String())))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	return values.NewGameFileTmpURL(tmpURL), nil
}
