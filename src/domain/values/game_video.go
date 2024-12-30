package values

import (
	"net/url"

	"github.com/google/uuid"
)

type (
	GameVideoID     uuid.UUID
	GameVideoType   int
	GameVideoTmpURL *url.URL
)

func NewGameVideoID() GameVideoID {
	return GameVideoID(uuid.New())
}

func NewGameVideoIDFromUUID(id uuid.UUID) GameVideoID {
	return GameVideoID(id)
}

const (
	GameVideoTypeMp4 GameVideoType = iota
	GameVideoTypeM4v
	GameVideoTypeMkv
)

func NewGameVideoTmpURL(tmpURL *url.URL) GameVideoTmpURL {
	return GameVideoTmpURL(tmpURL)
}
