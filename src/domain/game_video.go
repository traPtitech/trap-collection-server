package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

/*
	GameVideo
	ゲームの紹介映像。
*/
type GameVideo struct {
	id        values.GameVideoID
	videoType values.GameVideoType
	createdAt time.Time
}

func NewGameVideo(
	id values.GameVideoID,
	videoType values.GameVideoType,
	createdAt time.Time,
) *GameVideo {
	return &GameVideo{
		id:        id,
		videoType: videoType,
		createdAt: createdAt,
	}
}

func (v *GameVideo) GetID() values.GameVideoID {
	return v.id
}

func (v *GameVideo) GetType() values.GameVideoType {
	return v.videoType
}

func (v *GameVideo) GetCreatedAt() time.Time {
	return v.createdAt
}
