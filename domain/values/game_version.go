package values

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/mod/semver"
)

type (
	GameVersionID string
	// Semantic Versioning: https://semver.org/lang/ja/
	GameVersionName string
	GameVersionDescription string
	GameVersionCreatedAt time.Time
)

func NewGameVersionID() GameVersionID {
	return GameVersionID(uuid.New().String())
}

func NewGameVersionIDFromString(id string) (GameVersionID, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrInvalidFormat
	}

	return GameVersionID(id), nil
}

func NewGameVersionName(name string) (GameVersionName, error) {
	if !semver.IsValid(name) {
		return "", ErrInvalidFormat
	}

	return GameVersionName(name), nil
}

func NewGameVersionDescription(description string) (GameVersionDescription, error) {
	return GameVersionDescription(description), nil
}

func NewGameVersionCreatedAt(createdAt time.Time) (GameVersionCreatedAt, error) {
	return GameVersionCreatedAt(createdAt), nil
}
