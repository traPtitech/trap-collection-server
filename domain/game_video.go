package domain

import (
	"github.com/traPtitech/trap-collection-server/domain/values"
)

type GameVideo struct {
	id values.GameIntroductionID
	extension values.GameVideoExtension
}

func NewGameVideo(id values.GameIntroductionID, extension values.GameVideoExtension) *GameVideo {
	return &GameVideo{
		id: id,
		extension: extension,
	}
}

func (gv *GameVideo) GetID() values.GameIntroductionID {
	return gv.id
}

func (gv *GameVideo) GetExtension() values.GameVideoExtension {
	return gv.extension
}
