package service

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVersion interface {
	CreateGameVersion(ctx context.Context, gameID values.GameID, name values.GameVersionName, description values.GameVersionDescription) (*domain.GameVersion, error)
	GetGameVersions(ctx context.Context, gameID values.GameID) ([]*domain.GameVersion, error)
}
