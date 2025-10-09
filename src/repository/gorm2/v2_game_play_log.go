package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
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

	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}

	var endTime sql.NullTime
	if playLog.GetEndTime() != nil {
		endTime = sql.NullTime{
			Time:  *playLog.GetEndTime(),
			Valid: true,
		}
	}

	gamePlayLogTable := schema.GamePlayLogTable{
		ID:            uuid.UUID(playLog.GetID()),
		EditionID:     uuid.UUID(playLog.GetEditionID()),
		GameID:        uuid.UUID(playLog.GetGameID()),
		GameVersionID: uuid.UUID(playLog.GetGameVersionID()),
		StartTime:     playLog.GetStartTime(),
		EndTime:       endTime,
		CreatedAt:     playLog.GetCreatedAt(),
		UpdatedAt:     playLog.GetUpdatedAt(),
	}

	err = db.Create(&gamePlayLogTable).Error
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return repository.ErrDuplicatedUniqueKey
		}
		return fmt.Errorf("create game play log: %w", err)
	}
	return nil
}

func (g *GamePlayLogV2) GetGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) (*domain.GamePlayLog, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, err
	}

	var gamePlayLog schema.GamePlayLogTable //migrateではなくschemaに定義されている構造体を使う
	err = db.
		Where("id = ?", uuid.UUID(playLogID)). //playLogIDに合致したレコードを取得
		First(&gamePlayLog).Error              //1件を取得
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRecordNotFound
		}
		return nil, err
	}

	var endTime *time.Time // endTimeはNULL許容なのでポインタで扱う
	if gamePlayLog.EndTime.Valid {
		endTime = &gamePlayLog.EndTime.Time
	}

	return domain.NewGamePlayLog(
		values.GamePlayLogID(gamePlayLog.ID),
		values.LauncherVersionID(gamePlayLog.EditionID),
		values.GameID(gamePlayLog.GameID),
		values.GameVersionID(gamePlayLog.GameVersionID),
		gamePlayLog.StartTime,
		endTime,
		gamePlayLog.CreatedAt,
		gamePlayLog.UpdatedAt,
	), nil
}

func (g *GamePlayLogV2) UpdateGamePlayLogEndTime(_ context.Context, _ values.GamePlayLogID, _ time.Time) error {
	// TODO: interfaceのコメントを参考に実装を行う

	panic("not implemented")
}

func (g *GamePlayLogV2) GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error) {
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// start〜endの期間でフィルタリングする。
	// 統計データが存在しない場合でも空の統計を返すようにする。エラーは発生しない
	// ログはプレイ中でも含めたい カウント,プレイ時間にも含める

	db, err := g.db.getDB(ctx)
	if err != nil {
		err := fmt.Errorf("%s", "DB接続の取得に失敗")
		return nil, err
	}

	//Statsを取得して変数に入れる endtimeがNULLのものも含める
	stats := db.Where("game_id = ?", uuid.UUID(gameID)).
		Where("(start_time < ? AND end_time > ?) OR (start_time >= ? AND start_time < ? AND end_time IS NULL)", end, start, start, end)

	//gameVersionIDがnilでなければ絞り込み
	if gameVersionID != nil {
		stats = stats.Where("game_version_id = ?", uuid.UUID(*gameVersionID))
	}

	type hourlyResult struct {
		StartTime time.Time
		PlayCount int           //時間ごとのプレイ回数 あとで合計をとる
		PlayTime  sql.NullInt64 //時間ごとのプレイ時間(秒) あとでtime.Durationに変換して合計をとる
	}
	var hourlyResults []*hourlyResult //時間ごとのプレイ統計を入れるスライス
	//日付と時間を別々に取得して、start_timeを計算 play_countを計算 play_timeはifNullで計算
	err = stats.Model(&schema.GamePlayLogTable{}).
		Select("DATE_ADD(DATE(start_time), INTERVAL HOUR(start_time) HOUR) as start_time, COUNT(*) as play_count, SUM(TIMESTAMPDIFF(SECOND, start_time, IFNULL(end_time, ?))) as play_time", end).
		Group("DATE_FORMAT(start_time, '%Y-%m-%d %H:00:00')").
		Order("start_time").
		Scan(&hourlyResults).Error
	if err != nil {
		err := fmt.Errorf("%s", "時間ごとのプレイ統計の取得に失敗")
		return nil, err
	}

	//合計回数と時間も出す
	var totalPlayCount int          // 全体のプレイ回数
	var totalPlayTime time.Duration // 全体のプレイ時間
	hourlyStats := make([]*domain.HourlyPlayStats, 0, len(hourlyResults))

	for _, result := range hourlyResults {
		playTime := time.Duration(result.PlayTime.Int64) * time.Second //time.Durationはナノ秒単位なので秒に変換
		totalPlayCount += result.PlayCount                             //プレイ回数合計を計算
		totalPlayTime += playTime                                      //プレイ時間合計を計算

		stats := domain.NewHourlyPlayStats(
			result.StartTime,
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

func (g *GamePlayLogV2) GetEditionPlayStats(_ context.Context, _ values.LauncherVersionID, _, _ time.Time) (*domain.EditionPlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	panic("not implemented")
}
