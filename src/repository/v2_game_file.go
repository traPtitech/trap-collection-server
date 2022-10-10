package repository

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type GameFileV2 interface {
	// GetGameFiles
	// ゲームファイルのメタデータ一覧の取得。
	// 既にストレージに保存済みのファイルのみが取得できる。
	// ファイルの並び順はCreateAtの降順。
	GetGameFiles(ctx context.Context, fileIDs []values.GameFileID, lockType LockType) ([]*GameFileInfo, error)
}

type GameFileInfo struct {
	*domain.GameFile
	GameID values.GameID
}
