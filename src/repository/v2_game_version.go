package repository

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type (
	OptionURLLink = types.Option[values.GameURLLink]
)

type GameVersionV2 interface {
	// CreateGameVersion
	// ゲームバージョンの作成。
	CreateGameVersion(
		ctx context.Context,
		gameID values.GameID,
		imageID values.GameImageID,
		videoID values.GameVideoID,
		url OptionURLLink,
		fileIDs []values.GameFileID,
		version *domain.GameVersion,
	) error
	// GetGameVersions
	// ゲームに対応するゲームバージョンの一覧の取得。
	// 並び順はCreateAtの降順。
	// limitが0の場合、全てのゲームバージョンを取得する。
	GetGameVersions(
		ctx context.Context,
		gameID values.GameID,
		limit uint,
		offset uint,
		lockType LockType,
	) (uint, []*GameVersionInfo, error)
	// GetGameVersionsByIDs
	// ゲームバージョンIDの一覧からゲームバージョンの一覧を取得。
	// 並び順は引数の順番。
	GetGameVersionsByIDs(
		ctx context.Context,
		gameVersionIDs []values.GameVersionID,
		lockType LockType,
	) ([]*GameVersionInfoWithGameID, error)
	// GetGameVersionByID
	// ゲームバージョンIDからゲームバージョンを取得。
	// TODO: まだ実装が不正確(エラーを出さないために仮実装をおいている)ので，後で実装する
	GetGameVersionByID(
		ctx context.Context,
		gameVersionID values.GameVersionID,
		lockType LockType,
	) (*GameVersionInfoWithGameID, error)
	// GetLatestGameVersion
	// ゲームに対応する最新のゲームバージョンの取得。
	GetLatestGameVersion(
		ctx context.Context,
		gameID values.GameID,
		lockType LockType,
	) (*GameVersionInfo, error)
}

type GameVersionInfo struct {
	*domain.GameVersion
	ImageID values.GameImageID
	VideoID values.GameVideoID
	URL     OptionURLLink
	FileIDs []values.GameFileID
}

type GameVersionInfoWithGameID struct {
	*domain.GameVersion
	GameID  values.GameID
	ImageID values.GameImageID
	VideoID values.GameVideoID
	URL     OptionURLLink
	FileIDs []values.GameFileID
}
