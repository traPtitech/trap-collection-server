package domain

import (
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFile struct {
	id         values.GameFileID
	fileType   values.GameFileType
	entryPoint values.GameFileEntryPoint
	hash       values.GameFileHash
}

func NewGameFile(
	id values.GameFileID,
	fileType values.GameFileType,
	entryPoint values.GameFileEntryPoint,
	hash values.GameFileHash,
) *GameFile {
	return &GameFile{
		id:         id,
		fileType:   fileType,
		entryPoint: entryPoint,
		hash:       hash,
	}
}

func (gf *GameFile) GetID() values.GameFileID {
	return gf.id
}

func (gf *GameFile) GetFileType() values.GameFileType {
	return gf.fileType
}

func (gf *GameFile) GetEntryPoint() values.GameFileEntryPoint {
	return gf.entryPoint
}

func (gf *GameFile) GetHash() values.GameFileHash {
	return gf.hash
}
