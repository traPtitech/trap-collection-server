package values

import (
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"
)

type (
	GameID          uuid.UUID
	GameName        string
	GameDescription string
)

func NewGameID() GameID {
	return GameID(uuid.New())
}

func NewGameIDFromUUID(id uuid.UUID) GameID {
	return GameID(id)
}

func NewGameName(name string) GameName {
	return GameName(name)
}

var (
	ErrGameNameEmpty   = errors.New("game name is empty")
	ErrGameNameTooLong = errors.New("game name is too long")
)

func (gn GameName) Validate() error {
	// バージョン名は空ではない
	if len(gn) == 0 {
		return ErrGameNameEmpty
	}

	// バージョン名は32文字以内
	if utf8.RuneCountInString(string(gn)) > 32 {
		return ErrGameNameTooLong
	}

	return nil
}

func NewGameDescription(description string) GameDescription {
	return GameDescription(description)
}
