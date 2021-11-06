package local

import (
	"fmt"
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
