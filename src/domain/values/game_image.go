package values

import (
	"net/url"

	"github.com/google/uuid"
)

type (
	GameImageID     uuid.UUID
	GameImageType   int
	GameImageTmpURL *url.URL
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

func NewGameImageTmpURL(tmpURL *url.URL) GameImageTmpURL {
	return GameImageTmpURL(tmpURL)
}
