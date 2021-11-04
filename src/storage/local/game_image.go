package local

import (
	"fmt"
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
