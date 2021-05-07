package domain

import (
	"github.com/traPtitech/trap-collection-server/domain/values"
)

type GameImage struct {
	id values.GameIntroductionID
	extension values.GameImageExtension
}

func NewGameImage(id values.GameIntroductionID, extension values.GameImageExtension) *GameImage {
	return &GameImage{
		id: id,
		extension: extension,
	}
}

func (gi *GameImage) GetID() values.GameIntroductionID {
	return gi.id
}

func (gi *GameImage) GetExtension() values.GameImageExtension {
	return gi.extension
}
