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

func (g *GamePlayLogV2) GetGamePlayStats(_ context.Context, _ values.GameID, _ *values.GameVersionID, _, _ time.Time) (*domain.GamePlayStats, error) {
	// TODO: interfaceのコメントを参考に実装を行う
	// GetGamePlayStats
	// 指定されたゲームと期間のプレイ統計を取得する。
	// gameVersionIDがnilの場合、そのゲームのすべてのバージョンの統計を取得する。
	// start〜endの期間でフィルタリングする。
	// 統計データが存在しない場合でも空の統計を返すようにする。エラーは発生しない

	
	panic("not implemented")
}

func (g *GamePlayLogV2) GetEditionPlayStats(ctx context.Context, editionID values.LauncherVersionID, start, end time.Time) (*domain.EditionPlayStats, error) {
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
		values.NewLauncherVersionName(edition.Name),
		totalPlayCount,
		totalPlayTime,
		gameStats,
		hourlyStats,
	), nil
}
