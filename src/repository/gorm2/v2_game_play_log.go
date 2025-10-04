package gorm2

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
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
	// プレイ中でも含めたい カウントに含め プレイ時間にも含める
	// 時間はdb.goみたらAsiaだったのでJSTに揃えました

	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, err
	}

	//Statsを取得して変数に入れる endtimeがNULLのものも含める
	stats := db.Where("game_id = ?", uuid.UUID(gameID)).Where("start_time >= ? AND start_time < ?", start, end)

	//gameVersionIDがnilでなければ絞り込み
	if gameVersionID != nil {
		stats = stats.Where("game_version_id = ?", uuid.UUID(*gameVersionID))
	}

	type hourlyResult struct {
		StartTime string        //DATE_FORMATの形的にstringでないと受け取れない
		PlayCount int           //時間ごとのプレイ回数 あとで合計をとる
		PlayTime  sql.NullInt64 //時間ごとのプレイ時間(秒) あとでtime.Durationに変換して合計をとる
	}
	var hourlyResults []*hourlyResult //時間ごとのプレイ統計を入れるスライス
	//日付を時間単位で丸め込み play_countを計算 play_timeはifNullで計算
	err = stats.Model(&schema.GamePlayLogTable{}).
		Select("DATE_FORMAT(start_time, '%Y-%m-%d %H:00:00') as start_time, COUNT(*) as play_count, SUM(TIMESTAMPDIFF(SECOND, start_time, IFNULL(end_time, ?))) as play_time", end).
		Group("DATE_FORMAT(start_time, '%Y-%m-%d %H:00:00')").
		Order("start_time").
		Scan(&hourlyResults).Error
	if err != nil {
		return nil, err
	}

	jst, err := time.LoadLocation("Asia/Tokyo") //time.ParseInLocationで使うタイムゾーンを示すtime.Location型を作成
	if err != nil {
		return nil, err
	}

	//forで時間ごとの統計をtime.Timeに変換して、hourlyStatsを要求されている形に整理 ついでに合計回数と時間も出す
	var totalPlayCount int          // 全体のプレイ回数
	var totalPlayTime time.Duration // 全体のプレイ時間
	hourlyStats := make([]*domain.HourlyPlayStats, 0, len(hourlyResults))

	for _, result := range hourlyResults {
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", result.StartTime, jst) //jstにパース
		if err != nil {
			return nil, err
		}

		playTime := time.Duration(result.PlayTime.Int64) * time.Second //time.Durationはナノ秒単位なので秒に変換
		totalPlayCount += result.PlayCount                             //合計を計算
		totalPlayTime += playTime                                      //合計を計算

		stats := domain.NewHourlyPlayStats(
			startTime,
			result.PlayCount,
			playTime,
		)
		hourlyStats = append(hourlyStats, stats)
	}

	return domain.NewGamePlayStats(
		gameID,
		totalPlayCount,
		totalPlayTime,
		hourlyStats), nil

}

func (g *GamePlayLogV2) GetEditionPlayStats(_ context.Context, _ values.LauncherVersionID, _ time.Time, _ time.Time) (*domain.EditionPlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}
