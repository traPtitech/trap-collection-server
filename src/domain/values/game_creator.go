package values

import "github.com/google/uuid"

type GameCreatorID uuid.UUID

func NewGameCreatorID() GameCreatorID {
	return GameCreatorID(uuid.New())
}

type GameCreatorJobID uuid.UUID

func NewGameCreatorJobID() GameCreatorJobID {
	return GameCreatorJobID(uuid.New())
}

type GameCreatorJobDisplayName string

func NewGameCreatorJobDisplayName(name string) GameCreatorJobDisplayName {
	return GameCreatorJobDisplayName(name)
}
