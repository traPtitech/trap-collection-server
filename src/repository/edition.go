package repository

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Edition interface {
	// SaveEdition
	// エディションの保存。
	SaveEdition(ctx context.Context, edition *domain.LauncherVersion) error
	// UpdateEdition
	// エディションの更新。
	UpdateEdition(ctx context.Context, edition *domain.LauncherVersion) error
	// DeleteEdition
	// エディションの削除。
	DeleteEdition(ctx context.Context, editionID values.LauncherVersionID) error
	// GetEditions
	// エディションの一覧の取得。
	// 並び順はCreatedAtの降順。
	GetEditions(ctx context.Context, lockType LockType) ([]*domain.LauncherVersion, error)
	// GetEdition
	// エディションの取得。
	GetEdition(ctx context.Context, editionID values.LauncherVersionID, lockType LockType) (*domain.LauncherVersion, error)
	// UpdateEditionGameVersions
	// エディションに含まれるゲームバージョンの更新。
	UpdateEditionGameVersions(
		ctx context.Context,
		editionID values.LauncherVersionID,
		gameVersionIDs []values.GameVersionID,
	) error
	// GetEditionGameVersions
	// エディションに含まれるゲームバージョンの取得。
	// 並び順はCreatedAtの降順。
	GetEditionGameVersions(ctx context.Context, editionID values.LauncherVersionID, lockType LockType) ([]*GameVersionInfoWithGameID, error)
}
