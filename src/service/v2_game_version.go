package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

// mockgenがgenericsに対応するまでの暫定対応
// interface内にgenericsの構文が出なければ問題ないのを利用している
// ref: https://github.com/golang/mock/pull/640
// TODO: mockgenのv1.7.0がリリースされ次第削除する
type (
	OptionFileID  = types.Option[values.GameFileID]
	OptionURLLink = types.Option[values.GameURLLink]
)

// GameVersionV2
// ゲームバージョンの操作に関するサービス
type GameVersionV2 interface {
	// CreateGameVersion
	// ゲームバージョンの作成。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// 画像、動画、ファイルでそれぞれのIDに対応するものが存在しない場合、
	// ErrInvalidImageID、ErrInvalidVideoID、ErrInvalidFileIDを返す。
	// fileの種類が誤っている場合、ErrInvalidFileTypeを返す。
	// url、fileのいずれも空の場合、ErrNoAssetを返す。
	// gameIDとnameが同一の組み合わせが既に存在する場合、ErrDuplicateGameVersionを返す。
	CreateGameVersion(
		ctx context.Context,
		gameID values.GameID,
		name values.GameVersionName,
		description values.GameVersionDescription,
		imageID values.GameImageID,
		videoID values.GameVideoID,
		assets *Assets,
	) (*GameVersionInfo, error)
	// GetGameVersions
	// ゲームバージョン一覧の取得。
	// paramsがnilの場合、全てのゲームバージョンを取得する。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// Limitが0の場合、ErrInvalidLimitを返す。
	GetGameVersions(ctx context.Context, gameID values.GameID, params *GetGameVersionsParams) (uint, []*GameVersionInfo, error)
	// GetLatestGameVersion
	// 最新のゲームバージョンの取得。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// ゲームにバージョンが存在しない場合、ErrNoGameVersionを返す。
	GetLatestGameVersion(ctx context.Context, gameID values.GameID) (*GameVersionInfo, error)
}

// GetGameVersionsParams
// GetGameVersionsのパラメータ
type GetGameVersionsParams struct {
	Limit  uint
	Offset uint
}

type Assets struct {
	URL     OptionURLLink
	Windows OptionFileID
	Mac     OptionFileID
	Jar     OptionFileID
}

type GameVersionInfo struct {
	*domain.GameVersion
	Assets  *Assets
	ImageID values.GameImageID
	VideoID values.GameVideoID
}
