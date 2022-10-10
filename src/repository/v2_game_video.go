package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVideoV2 interface {
	// GetGameVideo
	// ゲーム動画のメタデータの取得。
	// 既にストレージに保存済みの動画のみが取得できる。
	GetGameVideo(ctx context.Context, gameVideoID values.GameVideoID, lockType LockType) (*GameVideoInfo, error)
}

type GameVideoInfo struct {
	*domain.GameVideo
	GameID values.GameID
}
