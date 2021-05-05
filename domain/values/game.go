package values

import (
	"time"

	"github.com/google/uuid"
)

type (
	GameID string
	GameName string
	GameDescription string
	GameCreatedAt time.Time
)

func NewGameID() GameID {
	return GameID(uuid.New().String())
}

func NewGameIDFromString(id string) (GameID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidFormat
	}

	return GameID(id), nil
}

func NewGameName(name string) (GameName, error) {
	if len(name) > 32 {
		return "", ErrTooLong
	}

	return GameName(name), nil
}

func NewGameDescription(description string) (GameDescription, error) {
	return GameDescription(description), nil
}

func NewGameCreatedAt(createdAt time.Time) (GameCreatedAt, error) {
	return GameCreatedAt(createdAt), nil
}
