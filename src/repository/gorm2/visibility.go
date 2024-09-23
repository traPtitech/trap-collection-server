package gorm2

import (
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

func convertVisibilityType(visibility values.GameVisibility) (string, error) {
	switch visibility {
	case values.GameVisibilityTypePublic:
		return migrate.GameVisibilityTypePublic, nil
	case values.GameVisibilityTypeLimited:
		return migrate.GameVisibilityTypeLimited, nil
	case values.GameVisibilityTypePrivate:
		return migrate.GameVisibilityTypePrivate, nil
	default:
		return "", fmt.Errorf("invalid visibility type: %d", visibility)
	}
}
