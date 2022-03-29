package local

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/traPtitech/trap-collection-server/src/config"
)

type DirectoryManager struct {
	rootPath string
}

func NewDirectoryManager(conf config.StorageLocal) (*DirectoryManager, error) {
	rootPath, err := conf.Path()
	if err != nil {
		return nil, fmt.Errorf("failed to get root path: %w", err)
	}

	return &DirectoryManager{
		rootPath: rootPath,
	}, nil
}

func (m *DirectoryManager) setupDirectory(directoryName string) (string, error) {
	directoryPath := path.Join(m.rootPath, directoryName)

	info, err := os.Stat(directoryPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(directoryPath, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		return directoryPath, nil
	}

	if !info.IsDir() {
		return "", fmt.Errorf("file is not directory")
	}

	return directoryPath, nil
}
