package repository

import (
	"context"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock -typed

type GamePlayLogV2 interface {
	// CreateGamePlayLog
	// 新しいゲームプレイログを作成する。
	// 引数として与えられたplayLogのIDが既に存在する場合，ErrDuplicatedUniqueKeyを返す
	CreateGamePlayLog(ctx context.Context, playLog *domain.GamePlayLog) error
	// GetGamePlayLog
	// 指定されたIDのゲームプレイログを取得する。
	// 該当するプレイログが存在しない場合，ErrRecordNotFoundを返す
	GetGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) (*domain.GamePlayLog, error)
	// UpdateGamePlayLogEndTime
	// 指定されたIDのゲームプレイログの終了時刻を更新する。
	// 該当するプレイログが存在しない場合，ErrNoRecordUpdatedを返す
	UpdateGamePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error
	// GetGamePlayStats
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// start〜endの期間でフィルタリングする。
	// 統計データが存在しない場合でも空の統計を返すようにする。エラーは発生しない
	GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error)
	// GetEditionPlayStats
	// 指定されたエディションと期間のプレイ統計を取得する。
	// start〜endの期間でフィルタリングする。
	// 統計データが存在しない場合でも空の統計を返すようにする。エラーは発生しない
	// editionNameも含めて返すため、editionsテーブルとのJOINが必要
	GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error)
}
