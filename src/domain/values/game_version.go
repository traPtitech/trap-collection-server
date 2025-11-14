package values

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/mod/semver"
)

type (
	GameVersionID uuid.UUID
	// セマンティックバージョニング
	GameVersionName        string
	GameVersionDescription string
)

func NewGameVersionID() GameVersionID {
	return GameVersionID(uuid.New())
}

func NewGameVersionIDFromUUID(id uuid.UUID) GameVersionID {
	return GameVersionID(id)
}

func NewGameVersionName(name string) GameVersionName {
	return GameVersionName(name)
}

var (
	ErrGameVersionNameInvalidSemanticVersion = errors.New("invalid semantic version")
)

func (gvn GameVersionName) Validate() error {
	if !semver.IsValid(string(gvn)) {
		return ErrGameVersionNameInvalidSemanticVersion
	}

	return nil
}

func NewGameVersionDescription(description string) GameVersionDescription {
	return GameVersionDescription(description)
}
