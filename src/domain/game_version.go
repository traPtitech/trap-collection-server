package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVersion struct {
	id          values.GameVersionID
	name        values.GameVersionName
	description values.GameVersionDescription
	createdAt   time.Time
}

func NewGameVersion(
	id values.GameVersionID,
	name values.GameVersionName,
	description values.GameVersionDescription,
	createdAt time.Time,
) *GameVersion {
	return &GameVersion{
		id:          id,
		name:        name,
		description: description,
		createdAt:   createdAt,
	}
}

func (gv *GameVersion) GetID() values.GameVersionID {
	return gv.id
}

func (gv *GameVersion) GetName() values.GameVersionName {
	return gv.name
}

func (gv *GameVersion) GetDescription() values.GameVersionDescription {
	return gv.description
}

func (gv *GameVersion) GetCreatedAt() time.Time {
	return gv.createdAt
}
