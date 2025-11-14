package gorm2

import (
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
)

func convertVisibilityType(visibility values.GameVisibility) (string, error) {
	switch visibility {
	case values.GameVisibilityTypePublic:
		return schema.GameVisibilityTypePublic, nil
	case values.GameVisibilityTypeLimited:
		return schema.GameVisibilityTypeLimited, nil
	case values.GameVisibilityTypePrivate:
		return schema.GameVisibilityTypePrivate, nil
	default:
		return "", fmt.Errorf("invalid visibility type: %d", visibility)
	}
}
