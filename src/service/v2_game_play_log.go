package service

import (
	"context"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock -typed

type GamePlayLogV2 interface {
	// CreatePlayLog
	// 新しいゲームプレイログを作成する。
	// エディションが存在しない場合、ErrInvalidEditionを返す。
	// ゲームが存在しない場合、ErrInvalidGameを返す。
	// ゲームバージョンが存在しない場合、ErrInvalidGameVersionを返す。
	CreatePlayLog(ctx context.Context, editionID values.EditionID, gameID values.GameID, gameVersionID values.GameVersionID, startTime time.Time) (*domain.GamePlayLog, error)
	// UpdatePlayLogEndTime
	// 指定されたプレイログの終了時刻を更新する。
	// プレイログが存在しない場合、ErrInvalidPlayLogIDを返す。
	// 終了時刻が開始時刻より前の場合、ErrInvalidEndTimeを返す。
	// プレイログがeditionIDとgameIDのペアに対応しない場合、ErrInvalidPlayLogEditionGamePairを返す。
	UpdatePlayLogEndTime(ctx context.Context, editionID values.EditionID, gameID values.GameID, playLogID values.GamePlayLogID, endTime time.Time) error
	// GetGamePlayStats
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// ゲームが存在しない場合、ErrInvalidGameを返す。
	// gameVersionIDが指定されており、そのゲームバージョンが存在しない場合、ErrInvalidGameVersionを返す。
	GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error)
	// GetEditionPlayStats
	// 指定されたエディションと期間のプレイ統計を取得する。
	// エディションが存在しない場合、ErrInvalidEditionを返す。
	GetEditionPlayStats(ctx context.Context, editionID values.EditionID, start, end time.Time) (*domain.EditionPlayStats, error)
	// DeleteGamePlayLog
	// 指定されたプレイログを削除する。
	// 条件に当てはまるプレイログが存在しない場合、ErrInvalidPlayLogIDを返す。
	DeleteGamePlayLog(ctx context.Context, editionID values.EditionID, gameID values.GameID, playLogID values.GamePlayLogID) error
	// DeleteLongLogs
	// 指定する時間(threshold)より長いプレイログを論理削除するrepositoryコードを呼びだす。。
	// threshold:　削除対象時間の閾値 ここservice層で決定する
	// ※これはCronで定期実行されています。
	DeleteLongLogs(ctx context.Context) error
}
