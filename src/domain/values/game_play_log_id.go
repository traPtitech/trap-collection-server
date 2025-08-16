package values

import (
	"github.com/google/uuid"
)

type GamePlayLogID uuid.UUID

func NewGamePlayLogID() GamePlayLogID {
	return GamePlayLogID(uuid.New())
}

func GamePlayLogIDFromUUID(id uuid.UUID) GamePlayLogID {
	return GamePlayLogID(id)
}

func (id GamePlayLogID) String() string {
	return uuid.UUID(id).String()
}

func (id GamePlayLogID) UUID() uuid.UUID {
	return uuid.UUID(id)
}