package values

import "github.com/google/uuid"

type (
	GameImageID   uuid.UUID
	GameImageType int
)

func NewGameImageID() GameImageID {
	return GameImageID(uuid.New())
}

func GameImageIDFromUUID(id uuid.UUID) GameImageID {
	return GameImageID(id)
}

const (
	GameImageTypeJpeg GameImageType = iota
	GameImageTypePng
	GameImageTypeGif
)
