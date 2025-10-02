package gorm2

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

func TestCreateGamePlayLog(t *testing.T) {
	t.Skip("実装していない関数なのでスキップします")
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description string
		playLog     *domain.GamePlayLog
		isErr       bool
		err         error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description: "TODO: add test case",
			playLog:     nil,
			isErr:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := gamePlayLogRepository.CreateGamePlayLog(ctx, testCase.playLog)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetGamePlayLog(t *testing.T) {
	t.Skip("実装していない関数なのでスキップします")
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description string
		playLogID   values.GamePlayLogID
		expectedLog *domain.GamePlayLog
		isErr       bool
		err         error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description: "TODO: add test case",
			playLogID:   values.NewGamePlayLogID(),
			expectedLog: nil,
			isErr:       true,
			err:         repository.ErrRecordNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			log, err := gamePlayLogRepository.GetGamePlayLog(ctx, testCase.playLogID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedLog, log)
			}
		})
	}
}

func TestUpdateGamePlayLogEndTime(t *testing.T) {
	t.Skip("実装していない関数なのでスキップします")
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description string
		playLogID   values.GamePlayLogID
		endTime     time.Time
		isErr       bool
		err         error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description: "TODO: add test case",
			playLogID:   values.NewGamePlayLogID(),
			endTime:     time.Now(),
			isErr:       true,
			err:         repository.ErrNoRecordUpdated,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			err := gamePlayLogRepository.UpdateGamePlayLogEndTime(ctx, testCase.playLogID, testCase.endTime)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// テスト用のランダムな日時を生成するヘルパー関数
func generateRandomEventTimes() (time.Time, time.Time, time.Time, time.Time) {
	// 1. 基準となる日時をランダムに生成
	// 例: 過去1年以内（365日 * 24時間）のどこかの時刻
	randomHoursAgo := time.Duration(rand.Intn(365*24)) * time.Hour
	baseTime := time.Now().Add(-randomHoursAgo)

	// 2. 各フィールドに加えるランダムな時間を生成
	// 必ず正の値になるように最小値を設定
	updateDuration := time.Duration(rand.Intn(60)) * time.Minute     // 0〜59分
	startDelay := time.Duration(rand.Intn(24)+1) * time.Hour         // 1〜24時間
	eventDuration := time.Duration(rand.Intn(5*60)+30) * time.Minute // 30分〜5時間29分

	// 3. 大小関係を維持して日時を組み立てる
	createdAt := baseTime
	updatedAt := createdAt.Add(updateDuration)
	startTime := updatedAt.Add(startDelay)
	endTime := startTime.Add(eventDuration)

	return createdAt, updatedAt, startTime, endTime
}

func TestGetGamePlayStats(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	edition1 := schema.EditionTable{
		ID:   uuid.New(),
		Name: "Test",
	}
	game1 := schema.GameTable2{
		ID:               uuid.New(),
		Name:             "Test",
		VisibilityTypeID: 1,
	}
	gameImage1 := schema.GameImageTable2{
		ID:          uuid.New(),
		GameID:      game1.ID,
		ImageTypeID: 1,
	}
	gameVideo1 := schema.GameVideoTable2{
		ID:          uuid.New(),
		GameID:      game1.ID,
		VideoTypeID: 1,
	}
	gameVersion1 := schema.GameVersionTable2{
		ID:          uuid.New(),
		GameID:      game1.ID,
		GameImageID: gameImage1.ID,
		GameVideoID: gameVideo1.ID,
		Name:        "Test",
		Description: "test",
	}

	createdAt1, updatedAt1, startTime1, endTime1 := generateRandomEventTimes()
	gamePlayLog1 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime1,
		EndTime:       sql.NullTime{Time: endTime1, Valid: true},
		CreatedAt:     createdAt1,
		UpdatedAt:     updatedAt1,
	}

	createdAt2, updatedAt2, startTime2, endTime2 := generateRandomEventTimes()
	gamePlayLog2 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime2,
		EndTime:       sql.NullTime{Time: endTime2, Valid: true},
		CreatedAt:     createdAt2,
		UpdatedAt:     updatedAt2,
	}

	createdAt3, updatedAt3, startTime3, endTime3 := generateRandomEventTimes()
	gamePlayLog3 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime3,
		EndTime:       sql.NullTime{Time: endTime3, Valid: true},
		CreatedAt:     createdAt3,
		UpdatedAt:     updatedAt3,
	}

	require.NoError(t, db.Create(&edition1).Error)
	require.NoError(t, db.Create(&game1).Error)
	require.NoError(t, db.Create(&gameImage1).Error)
	require.NoError(t, db.Create(&gameVideo1).Error)
	require.NoError(t, db.Create(&gameVersion1).Error)
	require.NoError(t, db.Create(&gamePlayLog1).Error)
	require.NoError(t, db.Create(&gamePlayLog2).Error)
	require.NoError(t, db.Create(&gamePlayLog3).Error)

	t.Cleanup(func() {
		ctx := context.Background()
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GamePlayLogTable{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVersionTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVideoTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameImageTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.EditionTable{}).Error)
	})

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description   string
		gameID        values.GameID
		gameVersionID values.GameVersionID
		start         time.Time
		end           time.Time
		expectedStats *domain.GamePlayStats
		isErr         bool
		err           error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description:   "普通に取得",
			gameID:        values.GameID(game1.ID),
			gameVersionID: values.GameVersionID(gameVersion1.ID),
			start:         time.Now().Add(-24 * time.Hour), // 24時間前
			end:           time.Now(),
			expectedStats: nil,
			isErr:         false,
		},
		{
			description:   "バージョンID無しで取得",
			gameID:        values.GameID(game1.ID),
			gameVersionID: nil,
			start:         time.Now().Add(-24 * time.Hour), // 24時間前
			end:           time.Now(),
			expectedStats: nil,
			isErr:         false,
		},
		{
			description:   "存在しないゲームIDで取得",
			gameID:        values.NewGameID(),
			gameVersionID: values.GameVersionID(gameVersion1.ID),
			start:         time.Now().Add(-24 * time.Hour), // 24時間前
			end:           time.Now(),
			expectedStats: nil,
			isErr:         false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			stats, err := gamePlayLogRepository.GetGamePlayStats(ctx, testCase.gameID, testCase.gameVersionID, testCase.start, testCase.end)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedStats, stats)
			}
		})
	}
}

func TestGetEditionPlayStats(t *testing.T) {
	t.Skip("実装していない関数なのでスキップします")
	t.Parallel()

	ctx := context.Background()

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description   string
		editionID     values.LauncherVersionID
		start         time.Time
		end           time.Time
		expectedStats *domain.EditionPlayStats
		isErr         bool
		err           error
	}

	// TODO: テストを実装する
	testCases := []test{
		{
			description:   "TODO: add test case",
			editionID:     values.NewLauncherVersionID(),
			start:         time.Now().Add(-24 * time.Hour),
			end:           time.Now(),
			expectedStats: nil,
			isErr:         false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			stats, err := gamePlayLogRepository.GetEditionPlayStats(ctx, testCase.editionID, testCase.start, testCase.end)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, testCase.err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedStats, stats)
			}
		})
	}
}
