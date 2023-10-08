package values

import (
	"errors"

	"github.com/google/uuid"
)

type (
	GameGenreID   uuid.UUID
	GameGenreName string
)

func NewGameGenreID() GameGenreID {
	return GameGenreID(uuid.New())
}

func GameGenreIDFromUUID(id uuid.UUID) GameGenreID {
	return GameGenreID(id)
}

func NewGameGenreName(name string) GameGenreName {
	return GameGenreName(name)
}

var (
	ErrGameGenreNameEmpty   = errors.New("game genre name must not be empty")
	ErrGameGenreNameTooLong = errors.New("game genre name must be no longer than 32 characters")
)

const gameGenreNameLimit = 32

// ジャンル名は0文字より長く32文字以下にする。
func (gn GameGenreName) Validate() error {
	if len([]rune(string(gn))) == 0 {
		return ErrGameGenreNameEmpty
	}
	if len([]rune(string(gn))) > gameGenreNameLimit {
		return ErrGameGenreNameTooLong
	}

	return nil
}
