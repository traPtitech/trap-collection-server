package values

import "github.com/google/uuid"

type (
	GameVideoID   uuid.UUID
	GameVideoType int
)

func NewGameVideoID() GameVideoID {
	return GameVideoID(uuid.New())
}

func NewGameVideoIDFromUUID(id uuid.UUID) GameVideoID {
	return GameVideoID(id)
}

const (
	GameVideoTypeMp4 GameVideoType = iota
)
