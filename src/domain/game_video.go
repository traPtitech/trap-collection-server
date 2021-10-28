package domain

import "github.com/traPtitech/trap-collection-server/src/domain/values"

/*
	GameVideo
	ゲームの紹介映像。
*/
type GameVideo struct {
	id        values.GameVideoID
	videoType values.GameVideoType
}

func NewGameVideo(
	id values.GameVideoID,
	videoType values.GameVideoType,
) *GameVideo {
	return &GameVideo{
		id:        id,
		videoType: videoType,
	}
}

func (v *GameVideo) GetID() values.GameVideoID {
	return v.id
}

func (v *GameVideo) GetType() values.GameVideoType {
	return v.videoType
}
