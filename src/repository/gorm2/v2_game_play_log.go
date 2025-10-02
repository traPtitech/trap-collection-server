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

func (g *GamePlayLogV2) CreateGamePlayLog(ctx context.Context, playLog *domain.GamePlayLog) error {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) GetGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) (*domain.GamePlayLog, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) UpdateGamePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error {
	// TODO: interfaceのコメントを参考に実装を行う
	// GetGamePlayStats
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// start〜endの期間でフィルタリングする。
	// 統計データが存在しない場合でも空の統計を返すようにする。エラーは発生しない

	
	panic("not implemented")
}

func (g *GamePlayLogV2) GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}

func (g *GamePlayLogV2) GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}
