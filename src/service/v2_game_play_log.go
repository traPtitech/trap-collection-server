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
	CreatePlayLog(ctx context.Context, editionID values.LauncherVersionID, gameID values.GameID, gameVersionID values.GameVersionID, startTime time.Time) (*domain.GamePlayLog, error)
	// UpdatePlayLogEndTime
	// 指定されたプレイログの終了時刻を更新する。
	// プレイログが存在しない場合、ErrInvalidPlayLogIDを返す。
	// 終了時刻が開始時刻より前の場合、ErrInvalidEndTimeを返す。
	UpdatePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error
	// GetGamePlayStats
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// ゲームが存在しない場合、ErrInvalidGameを返す。
	// gameVersionIDが指定されており、そのゲームバージョンが存在しない場合、ErrInvalidGameVersionを返す。
	GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error)
	// GetEditionPlayStats
	// 指定されたエディションと期間のプレイ統計を取得する。
	// エディションが存在しない場合、ErrInvalidEditionを返す。
	GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error)
}
