package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
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
		values.EditionID(gamePlayLog.EditionID),
		values.GameID(gamePlayLog.GameID),
		values.GameVersionID(gamePlayLog.GameVersionID),
		gamePlayLog.StartTime,
		endTime,
		gamePlayLog.CreatedAt,
		gamePlayLog.UpdatedAt,
	), nil
}

func (g *GamePlayLogV2) UpdateGamePlayLogEndTime(ctx context.Context, playLogID values.GamePlayLogID, endTime time.Time) error {

	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}

	result := db.
		Model(&schema.GamePlayLogTable{}).
		Where("id = ?", playLogID.UUID()).
		Update("end_time", endTime)

	err = result.Error

	if err != nil {
		return fmt.Errorf("update end_time: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}

func (g *GamePlayLogV2) GetGamePlayStats(ctx context.Context, gameID values.GameID, gameVersionID *values.GameVersionID, start, end time.Time) (*domain.GamePlayStats, error) {
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// start〜endの期間でフィルタリングする。
	// 統計データが存在しない場合でも空の統計を返すようにする。エラーは発生しない
	// ログはプレイ中でも含める カウント,プレイ時間にも含める
	// 00分を跨いだログは各時間帯に分割して集計する 総回数はダブってカウントしない

	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db: %w", err)
	}

	stats := db.
		Model(&schema.GamePlayLogTable{}).
		Where("game_id = ?", uuid.UUID(gameID))

	// gameVersionIDが指定されていれば絞り込み
	if gameVersionID != nil {
		stats = stats.Where("game_version_id = ?", uuid.UUID(*gameVersionID))
	}
	// start〜endの期間でフィルタリング
	stats = stats.Where("(start_time < ? AND end_time > ?) OR (start_time < ? AND end_time IS NULL)", end, start, end)

	var playLogs []*schema.GamePlayLogTable
	if err := stats.Find(&playLogs).Error; err != nil {
		return nil, fmt.Errorf("get game play logs: %w", err)
	}

	hours := int(end.Sub(start)/time.Hour) + 1 // ログの期間での00分始まりの時間帯の数
	hourlyStatsMap := make(map[time.Time]*domain.HourlyPlayStats, hours)

	for _, log := range playLogs {
		// 1.ログの整形
		// ログの期間 [logStart, logEnd) を計算 デフォルトはクエリの終了時刻
		logStart := log.StartTime
		logEnd := end

		// ログがクエリ終了時刻より前に終わっている場合は、その終了時刻を使う
		if log.EndTime.Valid && log.EndTime.Time.Before(end) {

			logEnd = log.EndTime.Time
		}

		// ログがクエリ開始時刻より前に始まっている場合は、クエリ開始時刻を使う
		if logStart.Before(start) {
			logStart = start
		}

		// 2. ログが終了するまで1時間ずつ処理 hourlyTimeはログの00分時間単位での開始時間
		for hourlyTime := logStart.Truncate(time.Hour); hourlyTime.Before(logEnd); hourlyTime = hourlyTime.Add(time.Hour) {

			// hourlyTimeでのプレイスタート時刻を計算
			hourlyStartTime := logStart
			if hourlyStartTime.Before(hourlyTime) {
				hourlyStartTime = hourlyTime
			}

			// hourlyTimeでのプレイ終了時刻を計算
			hourlyEndTime := logEnd
			if hourlyEndTime.After(hourlyTime.Add(time.Hour)) {
				hourlyEndTime = hourlyTime.Add(time.Hour)
			}

			playTime := hourlyEndTime.Sub(hourlyStartTime) //00分時間帯あたりのプレイ時間の計算

			stats, ok := hourlyStatsMap[hourlyTime] //hourlyTimeでキーを指定して取得してOKがでなければ新規作成、すでにあれば追加
			if !ok {
				hourlyStatsMap[hourlyTime] = domain.NewHourlyPlayStats(
					hourlyTime,
					1,
					playTime,
				)
			} else {
				hourlyStatsMap[hourlyTime] = domain.NewHourlyPlayStats(
					hourlyTime,
					stats.GetPlayCount()+1,
					stats.GetPlayTime()+playTime,
				)
			}
		}
	}

	// 3. 最終的な返却形式に整形する
	totalPlayCount := len(playLogs)
	var totalPlayTime time.Duration
	hourlyStats := make([]*domain.HourlyPlayStats, 0, len(hourlyStatsMap))

	for _, stats := range hourlyStatsMap {
		hourlyStats = append(hourlyStats, stats)
		totalPlayTime += stats.GetPlayTime()
	}

	// 時間順にソート
	slices.SortFunc(hourlyStats, func(a, b *domain.HourlyPlayStats) int {
		return a.GetStartTime().Compare(b.GetStartTime())
	})

	return domain.NewGamePlayStats(
		gameID,
		totalPlayCount,
		totalPlayTime,
		hourlyStats,
	), nil
}

func (g *GamePlayLogV2) GetEditionPlayStats(ctx context.Context, editionID values.EditionID, start, end time.Time) (*domain.EditionPlayStats, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("get db: %w", err)
	}

	var edition schema.EditionTable
	err = db.Where("id = ?", uuid.UUID(editionID)).First(&edition).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get edition: %w", err)
	}

	var playLogs []schema.GamePlayLogTable
	err = db.Model(&schema.GamePlayLogTable{}).
		Where("edition_id = ?", uuid.UUID(editionID)).
		Where("start_time < ?", end).
		Where("(end_time > ? OR end_time IS NULL)", start).
		Order("start_time").
		Find(&playLogs).Error
	if err != nil {
		return nil, fmt.Errorf("get play logs: %w", err)
	}

	hourlyStatsMap := make(map[time.Time]struct {
		playTime  time.Duration
		playCount int
	})

	gameStatsMap := make(map[uuid.UUID]struct {
		playTime  time.Duration
		playCount int
	})

	var totalPlayCount int
	var totalPlayTime time.Duration

	for _, playLog := range playLogs {
		logStart := playLog.StartTime
		var logEnd time.Time
		if playLog.EndTime.Valid {
			logEnd = playLog.EndTime.Time
		} else {
			logEnd = end
		}

		// 検索期間内のプレイログを取得する
		if logStart.Before(start) {
			logStart = start
		}
		if logEnd.After(end) {
			logEnd = end
		}

		playDuration := logEnd.Sub(logStart)
		totalPlayCount++
		totalPlayTime += playDuration

		gameStats := gameStatsMap[playLog.GameID]
		gameStats.playCount++
		gameStats.playTime += playDuration
		gameStatsMap[playLog.GameID] = gameStats

		isFirstHour := true

		for hourlyRangeStart := time.Date(logStart.Year(), logStart.Month(), logStart.Day(), logStart.Hour(), 0, 0, 0, logStart.Location()); hourlyRangeStart.Before(logEnd); hourlyRangeStart = hourlyRangeStart.Add(time.Hour) {
			nextHour := hourlyRangeStart.Add(time.Hour)

			playStart := logStart
			if playStart.Before(hourlyRangeStart) {
				playStart = hourlyRangeStart
			}

			playEnd := logEnd
			if playEnd.After(nextHour) {
				playEnd = nextHour
			}

			if playStart.Before(playEnd) {
				playTimeInHour := playEnd.Sub(playStart)

				hourlyStats := hourlyStatsMap[hourlyRangeStart]
				hourlyStats.playTime += playTimeInHour
				if isFirstHour {
					hourlyStats.playCount++
				}
				hourlyStatsMap[hourlyRangeStart] = hourlyStats
			}

			isFirstHour = false
		}
	}

	hourlyStats := make([]*domain.HourlyPlayStats, 0, len(hourlyStatsMap))
	for hourTime, stats := range hourlyStatsMap {
		hourlyStats = append(hourlyStats, domain.NewHourlyPlayStats(
			hourTime,
			stats.playCount,
			stats.playTime,
		))
	}

	slices.SortFunc(hourlyStats, func(a, b *domain.HourlyPlayStats) int {
		return a.GetStartTime().Compare(b.GetStartTime())
	})

	// ゲーム別統計をスライスに変換
	gameStats := make([]*domain.GamePlayStatsInEdition, 0, len(gameStatsMap))
	for gameID, stats := range gameStatsMap {
		gameStats = append(gameStats, domain.NewGamePlayStatsInEdition(
			values.GameID(gameID),
			stats.playCount,
			stats.playTime,
		))
	}

	return domain.NewEditionPlayStats(
		editionID,
		values.NewEditionName(edition.Name),
		totalPlayCount,
		totalPlayTime,
		gameStats,
		hourlyStats,
	), nil
}

func (g *GamePlayLogV2) DeleteGamePlayLog(ctx context.Context, playLogID values.GamePlayLogID) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}

	result := db.Unscoped().
		Delete(&schema.GamePlayLogTable{}, "id = ?", uuid.UUID(playLogID))
	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}
	if result.Error != nil {
		return fmt.Errorf("delete game play log: %w", result.Error)
	}

	return nil
}

func (g *GamePlayLogV2) DeleteLongLogs(ctx context.Context, threshold time.Duration) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}

	// thresholdより前のstart_timeを持つプレイ中ログを論理削除
	deleteTime := time.Now().Add(-threshold)
	err = db.
		Where("end_time IS NULL").
		Where("start_time < ?", deleteTime).
		Delete(&schema.GamePlayLogTable{}).Error
	if err != nil {
		return fmt.Errorf("delete long play logs: %w", err)
	}

	return nil
}
