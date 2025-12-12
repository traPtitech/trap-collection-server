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
	GetEditionPlayStats(ctx context.Context, editionID values.EditionID, start, end time.Time) (*domain.EditionPlayStats, error)
	// DeleteGamePlayLog
	// 指定されたプレイログを削除する。
	// 条件に当てはまるプレイログが存在しない場合、ErrNoRecordDeletedを返す。
	DeleteGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) error
	//　DeleteLongLogs
	// 指定された時間(threshold)より長いプレイログを論理削除する。
	// threshold: 削除対象とする閾値（この時間より長いログを削除する）
	// ※これはCronで定期実行している関数です。
	DeleteLongLogs(ctx context.Context, threshold time.Duration) error
}
