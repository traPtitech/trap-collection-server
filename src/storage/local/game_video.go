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

type GameVideo struct {
	videoRootPath    string
	directoryManager *DirectoryManager
}

func NewGameVideo(directoryManager *DirectoryManager) (*GameVideo, error) {
	videoRootPath, err := directoryManager.setupDirectory("videos")
	if err != nil {
		return nil, fmt.Errorf("failed to setup directory: %w", err)
	}

	return &GameVideo{
		videoRootPath:    videoRootPath,
		directoryManager: directoryManager,
	}, nil
}

func (gv *GameVideo) SaveGameVideo(ctx context.Context, reader io.Reader, video *domain.GameVideo) error {
	videoPath := path.Join(gv.videoRootPath, uuid.UUID(video.GetID()).String())

	_, err := os.Stat(videoPath)
	if err == nil {
		return storage.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	f, err := os.Create(videoPath)
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

func (gv *GameVideo) GetGameVideo(ctx context.Context, writer io.Writer, video *domain.GameVideo) error {
	videoPath := path.Join(gv.videoRootPath, uuid.UUID(video.GetID()).String())

	f, err := os.Open(videoPath)
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
