package local

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/traPtitech/trap-collection-server/pkg/common"
)

type DirectoryManager struct {
	rootPath string
}

func NewDirectoryManager(rootPath common.FilePath) *DirectoryManager {
	return &DirectoryManager{
		rootPath: string(rootPath),
	}
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
