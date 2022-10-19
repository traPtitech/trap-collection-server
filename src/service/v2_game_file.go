package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFileV2 interface {
	// SaveGameFile
	// ゲームファイルの保存。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	SaveGameFile(ctx context.Context, reader io.Reader, gameID values.GameID, fileType values.GameFileType, entryPoint values.GameFileEntryPoint) (*domain.GameFile, error)
	// GetGameFile
	// ゲームファイル一覧の取得。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	GetGameFiles(ctx context.Context, gameID values.GameID) ([]*domain.GameFile, error)
	// GetGameFile
	// ゲームファイルの一時的(1分間)に有効なurlを返す。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// ゲームファイルIDに対応するゲームファイルが存在しない、
	// もしくは存在しても紐づくゲームのゲームIDが異なる場合、ErrInvalidGameFileIDを返す。
	GetGameFile(ctx context.Context, gameID values.GameID, fileID values.GameFileID) (values.GameFileTmpURL, error)
	// GetGameFileMeta
	// ゲームファイルのメタデータの取得。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// ゲームファイルIDに対応するゲームファイルが存在しない場合、ErrInvalidGameFileIDを返す。
	GetGameFileMeta(ctx context.Context, gameID values.GameID, fileID values.GameFileID) (*domain.GameFile, error)
}
