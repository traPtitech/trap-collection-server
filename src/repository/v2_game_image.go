package repository

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// GameImageV2
// ゲーム画像のメタデータの保存・取得。
// ストレージに保存済みの画像のメタデータのみが見えるように使う。
type GameImageV2 interface {
	// SaveGameImage
	// ゲーム画像のメタデータの保存。
	// 他のスレッドからはストレージに保存済みの画像のみが見えるようにトランザクションをとるように注意する。
	SaveGameImage(ctx context.Context, gameID values.GameID, image *domain.GameImage) error
	// GetGameImage
	// ゲーム画像のメタデータの取得。
	// 既にストレージに保存済みの画像のみが取得できる。
	GetGameImage(ctx context.Context, gameImageID values.GameImageID, lockType LockType) (*GameImageInfo, error)
	// GetGameImages
	// ゲームに対応するゲーム画像のメタデータ一覧の取得。
	// 既にストレージに保存済みの画像のみが取得できる。
	// 画像の並び順はCreateAtの降順。
	GetGameImages(ctx context.Context, gameID values.GameID, lockType LockType) ([]*domain.GameImage, error)
}

type GameImageInfo struct {
	*domain.GameImage
	GameID values.GameID
}
