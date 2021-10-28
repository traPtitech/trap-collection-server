package domain

import "github.com/traPtitech/trap-collection-server/src/domain/values"

/*
	GameImage
	ゲームの紹介画像。
*/
type GameImage struct {
	id        values.GameImageID
	imageType values.GameImageType
}

func NewGameImage(
	id values.GameImageID,
	imageType values.GameImageType,
) *GameImage {
	return &GameImage{
		id:        id,
		imageType: imageType,
	}
}

func (gi *GameImage) GetID() values.GameImageID {
	return gi.id
}

func (gi *GameImage) GetType() values.GameImageType {
	return gi.imageType
}
