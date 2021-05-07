package domain

import "github.com/traPtitech/trap-collection-server/domain/values"

type GameVersion struct {
	id values.GameVersionID
	name values.GameVersionName
	description values.GameVersionDescription
	createdAt values.GameVersionCreatedAt
}

func NewGameVersion(id values.GameVersionID, name values.GameVersionName, description values.GameVersionDescription, createdAt values.GameVersionCreatedAt) *GameVersion {
	return &GameVersion{
		id: id,
		name: name,
		description: description,
		createdAt: createdAt,
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

func (gv *GameVersion) GetCreatedAt() values.GameVersionCreatedAt {
	return gv.createdAt
}
