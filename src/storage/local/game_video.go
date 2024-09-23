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

func (gv *GameVideo) SaveGameVideo(_ context.Context, reader io.Reader, videoID values.GameVideoID) error {
	videoPath := path.Join(gv.videoRootPath, uuid.UUID(videoID).String())

	_, err := os.Stat(videoPath)
	if err == nil {
		return storage.ErrAlreadyExists
	}
	if !errors.Is(err, fs.ErrNotExist) {
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

func (gv *GameVideo) GetTempURL(_ context.Context, video *domain.GameVideo, _ time.Duration) (values.GameVideoTmpURL, error) {
	// 正しいURLにはならないが、開発環境用のmockのため妥協する
	tmpURL, err := url.Parse(fmt.Sprintf("file://%s", path.Join(gv.videoRootPath, uuid.UUID(video.GetID()).String())))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	return values.NewGameVideoTmpURL(tmpURL), nil
}
