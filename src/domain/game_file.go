package domain

import (
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFile struct {
	id         values.GameFileID
	entryPoint values.GameFileEntryPoint
	hash       values.GameFileHash
}

func NewGameFile(
	id values.GameFileID,
	entryPoint values.GameFileEntryPoint,
	hash values.GameFileHash,
) *GameFile {
	return &GameFile{
		id:         id,
		entryPoint: entryPoint,
		hash:       hash,
	}
}

func (gf *GameFile) GetID() values.GameFileID {
	return gf.id
}

func (gf *GameFile) GetEntryPoint() values.GameFileEntryPoint {
	return gf.entryPoint
}

func (gf *GameFile) GetHash() values.GameFileHash {
	return gf.hash
}
