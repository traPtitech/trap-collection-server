package repository

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVersion interface {
	CreateGameVersion(ctx context.Context, gameID values.GameID, version *domain.GameVersion) error
	// GetGameVersions CreatedAtで降順にソートしたGameVersionのリストを取得
	GetGameVersions(ctx context.Context, gameID values.GameID) ([]*domain.GameVersion, error)
	GetLatestGameVersion(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.GameVersion, error)
	GetLatestGameVersionsByGameIDs(ctx context.Context, gameIDs []values.GameID, lockType LockType) (map[values.GameID]*domain.GameVersion, error)
}
