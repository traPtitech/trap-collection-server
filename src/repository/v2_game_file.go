package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// GameFileV2
// ゲームファイルのメタデータの保存・取得。
// ストレージに保存済みのファイルのメタデータのみが見えるように使う。
type GameFileV2 interface {
	// SaveGameFile
	// ゲームファイルのメタデータの保存。
	// 他のスレッドからはストレージに保存済みのファイルのみが見えるようにトランザクションをとるように注意する。
	SaveGameFile(ctx context.Context, gameID values.GameID, file *domain.GameFile) error
	// GetGameFile
	// ゲームファイルのメタデータの取得。
	// 既にストレージに保存済みのファイルのみが取得できる。
	GetGameFile(ctx context.Context, gameFileID values.GameFileID, lockType LockType) (*GameFileInfo, error)
	// GetGameFiles
	// ゲームに対応するゲームファイルのメタデータ一覧の取得。
	// 既にストレージに保存済みのファイルのみが取得できる。
	// ファイルの並び順はCreateAtの降順。
	GetGameFiles(ctx context.Context, gameID values.GameID, lockType LockType, fileTypes []values.GameFileType) ([]*domain.GameFile, error)
	// GetGameFilesWithoutTypes
	// ゲームファイルのメタデータ一覧の取得。
	// 既にストレージに保存済みのファイルのみが取得できる。
	// ファイルの並び順はCreateAtの降順。
	GetGameFilesWithoutTypes(ctx context.Context, fileIDs []values.GameFileID, lockType LockType) ([]*GameFileInfo, error)
}

type GameFileInfo struct {
	*domain.GameFile
	GameID values.GameID
}
