package repository

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type ProductKey interface {
	// SaveProductKeys
	// プロダクトキーの保存。
	SaveProductKeys(ct context.Context, editionID values.EditionID, productKeys []*domain.LauncherUser) error
	// UpdateProductKey
	// プロダクトキーの更新。
	UpdateProductKey(ctx context.Context, productKey *domain.LauncherUser) error
	// GetProductKeys
	// プロダクトキー一覧の取得。
	// ステータスに関わらず取得可能。
	GetProductKeys(ctx context.Context, editionID values.EditionID, statuses []values.LauncherUserStatus, lockType LockType) ([]*domain.LauncherUser, error)
	// GetProductKey
	// プロダクトキーの取得。
	// ステータスに関わらず取得可能。
	GetProductKey(ctx context.Context, productKeyID values.LauncherUserID, lockType LockType) (*domain.LauncherUser, error)
	// GetProductKeyByKey
	// キーからのプロダクトキーの取得。
	// ステータスに関わらず取得可能。
	GetProductKeyByKey(ctx context.Context, key values.LauncherUserProductKey) (*domain.LauncherUser, error)
}
