package local

import "fmt"

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
