package repository

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// GameVideoV2
// ゲーム動画のメタデータの保存・取得。
// ストレージに保存済みの動画のメタデータのみが見えるように使う。
type GameVideoV2 interface {
	// SaveGameVideo
	// ゲーム動画のメタデータの保存。
	// 他のスレッドからはストレージに保存済みの動画のみが見えるようにトランザクションをとるように注意する。
	SaveGameVideo(ctx context.Context, gameID values.GameID, video *domain.GameVideo) error
	// GetGameVideo
	// ゲーム動画のメタデータの取得。
	// 既にストレージに保存済みの動画のみが取得できる。
	GetGameVideo(ctx context.Context, gameVideoID values.GameVideoID, lockType LockType) (*GameVideoInfo, error)
	// GetGameVideos
	// ゲームに対応するゲーム動画のメタデータ一覧の取得。
	// 既にストレージに保存済みの動画のみが取得できる。
	// 動画の並び順はCreateAtの降順。
	GetGameVideos(ctx context.Context, gameID values.GameID, lockType LockType) ([]*domain.GameVideo, error)
}

type GameVideoInfo struct {
	*domain.GameVideo
	GameID values.GameID
}
