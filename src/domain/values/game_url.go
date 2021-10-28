package values

import (
	"net/url"

	"github.com/google/uuid"
)

type (
	GameURLID   uuid.UUID
	GameURLLink *url.URL
)

func NewGameURLID() GameURLID {
	return GameURLID(uuid.New())
}

func NewGameURLIDFromUUID(id uuid.UUID) GameURLID {
	return GameURLID(id)
}

func NewGameURLLink(link *url.URL) GameURLLink {
	return GameURLLink(link)
}
