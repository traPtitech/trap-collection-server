package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVersion interface {
	CreateGameVersion(ctx context.Context, gameID values.GameID, name values.GameVersionName, description values.GameVersionDescription) (*domain.GameVersion, error)
	GetGameVersions(ctx context.Context, gameID values.GameID) ([]*domain.GameVersion, error)
}
