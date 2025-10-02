package gorm2

import (
	"context"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GamePlayLogV2 struct {
	db *DB
}

func NewGamePlayLogV2(db *DB) *GamePlayLogV2 {
	return &GamePlayLogV2{
		db: db,
	}
}

func (g *GamePlayLogV2) CreateGamePlayLog(_ context.Context, _ *domain.GamePlayLog) error {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) GetGamePlayLog(_ context.Context, _ values.GamePlayLogID) (*domain.GamePlayLog, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) UpdateGamePlayLogEndTime(_ context.Context, _ values.GamePlayLogID, _ time.Time) error {
	// TODO: interfaceのコメントを参考に実装を行う

	panic("not implemented")
}

func (g *GamePlayLogV2) GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	// GetGamePlayStats
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// start〜endの期間でフィルタリングする。
	// 統計データが存在しない場合でも空の統計を返すようにする。エラーは発生しない

	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, err
	}

	stats := db.Where("game_id = ?", gameID).Where("start_time >= ? AND start_time < ?", start, end)

	if gameVersionID != nil {
		stats = stats.Where("game_version_id = ?", *gameVersionID)
	}

	var totalPlayCount int
	var totalPlayTime time.Duration

	err = stats.Select("COUNT(*)").Scan(&totalPlayCount).Error
	if err != nil {
		return nil, err
	}

	err = stats.Where("end_time IS NOT NULL").Select("SUM(TIMESTAMPDIFF(SECOND, start_time, end_time))").Scan(&totalPlayTime).Error
	if err != nil {
		return nil, err
	}

	

	return domain.NewGamePlayStats(gameID, totalPlayCount, totalPlayTime, []*domain.HourlyPlayStats{}), nil

}

func (g *GamePlayLogV2) GetEditionPlayStats(_ context.Context, _ values.LauncherVersionID, _ time.Time, _ time.Time) (*domain.EditionPlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}
