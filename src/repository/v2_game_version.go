package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

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
