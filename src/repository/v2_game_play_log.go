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
	CreateGamePlayLog(ctx context.Context, playLog *domain.GamePlayLog) error
	// GetGamePlayLog
	// 指定されたIDのゲームプレイログを取得する。
	GetGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) (*domain.GamePlayLog, error)
	// UpdateGamePlayLogEndTime
	// 指定されたIDのゲームプレイログの終了時刻を更新する。
	UpdateGamePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error
	// GetGamePlayStats
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// start〜endの期間でフィルタリングする。
	GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error)
	// GetEditionPlayStats
	// 指定されたエディションと期間のプレイ統計を取得する。
	// start〜endの期間でフィルタリングする。
	GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error)
}
