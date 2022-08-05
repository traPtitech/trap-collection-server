package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// GameImage
// ゲームの紹介画像。
type GameImage struct {
	id        values.GameImageID
	imageType values.GameImageType
	createdAt time.Time
}

func NewGameImage(
	id values.GameImageID,
	imageType values.GameImageType,
	createdAt time.Time,
) *GameImage {
	return &GameImage{
		id:        id,
		imageType: imageType,
		createdAt: createdAt,
	}
}

func (gi *GameImage) GetID() values.GameImageID {
	return gi.id
}

func (gi *GameImage) GetType() values.GameImageType {
	return gi.imageType
}

func (gi *GameImage) GetCreatedAt() time.Time {
	return gi.createdAt
}
