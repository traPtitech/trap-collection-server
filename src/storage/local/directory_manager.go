package local

import (
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
