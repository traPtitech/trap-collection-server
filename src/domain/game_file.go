package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFile struct {
	id         values.GameFileID
	fileType   values.GameFileType
	entryPoint values.GameFileEntryPoint
	hash       values.GameFileHash
	createdAt  time.Time
}

func NewGameFile(
	id values.GameFileID,
	fileType values.GameFileType,
	entryPoint values.GameFileEntryPoint,
	hash values.GameFileHash,
	createdAt time.Time,
) *GameFile {
	return &GameFile{
		id:         id,
		fileType:   fileType,
		entryPoint: entryPoint,
		hash:       hash,
		createdAt:  createdAt,
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

func (gf *GameFile) GetCreatedAt() time.Time {
	return gf.createdAt
}
