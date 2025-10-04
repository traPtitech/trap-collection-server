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

func TestGetGamePlayStats(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	testGameID := values.NewGameID()

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

	gameVersion2 := schema.GameVersionTable2{
		ID:          uuid.New(),
		GameID:      game1.ID,
		GameImageID: gameImage1.ID,
		GameVideoID: gameVideo1.ID,
		Name:        "Test2",
		Description: "test2",
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	baseTime := time.Date(2025, 9, 3, 0, 0, 0, 0, jst) // JSTの時間で設定

	// gemVersion1のログを5つ作成
	// gamePlayLog1: 15時台のログ 10分
	startTime1 := baseTime.Add(15 * time.Hour)   // 2025-09-03 15:00:00
	endTime1 := startTime1.Add(10 * time.Minute) // 2025-09-03 15:10:00
	gamePlayLog1 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime1,
		EndTime:       sql.NullTime{Time: endTime1, Valid: true},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// gamePlayLog2: 15時台のログ 20分
	startTime2 := baseTime.Add(15*time.Hour + 30*time.Minute) // 2025-09-03 15:30:00
	endTime2 := startTime2.Add(20 * time.Minute)              // 2025-09-03 15:50:00
	gamePlayLog2 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime2,
		EndTime:       sql.NullTime{Time: endTime2, Valid: true},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// gamePlayLog3: 16時台のログ 30分
	startTime3 := baseTime.Add(16 * time.Hour)   // 2025-09-03 16:00:00
	endTime3 := startTime3.Add(30 * time.Minute) // 2025-09-03 16:30:00
	gamePlayLog3 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime3,
		EndTime:       sql.NullTime{Time: endTime3, Valid: true},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// gamePlayLog4: 17時台のログ 40分
	startTime4 := baseTime.Add(17 * time.Hour)   // 2025-09-03 17:00:00
	endTime4 := startTime4.Add(40 * time.Minute) // 2025-09-03 17:40:00
	gamePlayLog4 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime4,
		EndTime:       sql.NullTime{Time: endTime4, Valid: true},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// gamePlayLog5: 18時台のまだ終わっていないログ
	startTime5 := baseTime.Add(18 * time.Hour) // 2025-09-03 18:00:00
	gamePlayLog5 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     startTime5,
		EndTime:       sql.NullTime{Time: time.Time{}, Valid: false},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	//gameVersion2のログを1つ作成 30分
	startTime6 := baseTime.Add(15 * time.Hour)   // 2025-09-03 15:00:00
	endTime6 := startTime6.Add(30 * time.Minute) // 2025-09-03 15:30:00
	gamePlayLog6 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion2.ID,
		StartTime:     startTime6,
		EndTime:       sql.NullTime{Time: endTime6, Valid: true},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	require.NoError(t, db.Create(&edition1).Error)
	require.NoError(t, db.Create(&game1).Error)
	require.NoError(t, db.Create(&gameImage1).Error)
	require.NoError(t, db.Create(&gameVideo1).Error)
	require.NoError(t, db.Create(&gameVersion1).Error)
	require.NoError(t, db.Create(&gameVersion2).Error)
	require.NoError(t, db.Create(&gamePlayLog1).Error)
	require.NoError(t, db.Create(&gamePlayLog2).Error)
	require.NoError(t, db.Create(&gamePlayLog3).Error)
	require.NoError(t, db.Create(&gamePlayLog4).Error)
	require.NoError(t, db.Create(&gamePlayLog5).Error)
	require.NoError(t, db.Create(&gamePlayLog6).Error)

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
		gameVersionID *values.GameVersionID
		start         time.Time
		end           time.Time
		expectedStats *domain.GamePlayStats
		isErr         bool
		err           error
	}

	gameVersion1ID := values.GameVersionID(gameVersion1.ID)

	testCases := []test{
		{
			description:   "gameVersion1を指定して取得",
			gameID:        values.GameID(game1.ID),
			gameVersionID: &gameVersion1ID,
			start:         baseTime.Add(15 * time.Hour), // 2025-10-03 15:00:00
			end:           baseTime.Add(17 * time.Hour), // 2025-10-03 17:00:00
			expectedStats: domain.NewGamePlayStats(
				values.GameID(game1.ID),
				3,              // totalPlayCount
				60*time.Minute, // totalPlayTime
				[]*domain.HourlyPlayStats{
					domain.NewHourlyPlayStats(
						baseTime.Add(15*time.Hour), // 15時台の開始時刻
						2,                          // 15時台のプレイ回数
						30*time.Minute,             // 15時台のプレイ時間
					),
					domain.NewHourlyPlayStats(
						baseTime.Add(16*time.Hour), // 16時台の開始時刻
						1,                          // 16時台のプレイ回数
						30*time.Minute,             // 16時台のプレイ時間
					),
				},
			),
			isErr: false,
		},
		{
			description:   "gameVersion1のプレイ中のログを含めて取得",
			gameID:        values.GameID(game1.ID),
			gameVersionID: &gameVersion1ID,
			start:         baseTime.Add(16 * time.Hour),                // 2025-10-03 16:00:00
			end:           baseTime.Add(18*time.Hour + 20*time.Minute), // 2025-10-03 18:20:00
			expectedStats: domain.NewGamePlayStats(
				values.GameID(game1.ID),
				3,
				90*time.Minute,
				[]*domain.HourlyPlayStats{
					domain.NewHourlyPlayStats(
						baseTime.Add(16*time.Hour), // 16時台の開始時刻
						1,                          // 16時台のプレイ回数
						30*time.Minute,             // 16時台のプレイ時間
					),
					domain.NewHourlyPlayStats(
						baseTime.Add(17*time.Hour), // 17時台の開始時刻
						1,                          // 17時台のプレイ回数
						40*time.Minute,             // 17時台のプレイ時間
					),
					domain.NewHourlyPlayStats(
						baseTime.Add(18*time.Hour), // 18時台の開始時刻
						1,                          // 18時台のプレイ回数
						20*time.Minute,             // 18時台のプレイ時間 (プレイ中の分も含む)
					),
				},
			),
			isErr: false,
		},
		{
			description:   "バージョンID無しで取得 (全バージョン集計)",
			gameID:        values.GameID(game1.ID),
			gameVersionID: nil,
			start:         baseTime,
			end:           baseTime.Add(18*time.Hour + 20*time.Minute),
			expectedStats: domain.NewGamePlayStats(
				values.GameID(game1.ID),
				6,
				150*time.Minute,
				[]*domain.HourlyPlayStats{
					domain.NewHourlyPlayStats(
						baseTime.Add(15*time.Hour), // 15時台
						3,
						60*time.Minute,
					),
					domain.NewHourlyPlayStats(
						baseTime.Add(16*time.Hour), // 16時台
						1,
						30*time.Minute,
					),
					domain.NewHourlyPlayStats(
						baseTime.Add(17*time.Hour), // 17時台
						1,
						40*time.Minute,
					),
					domain.NewHourlyPlayStats(
						baseTime.Add(18*time.Hour), // 18時台
						1,
						20*time.Minute,
					),
				},
			),
			isErr: false,
		},
		{
			description:   "存在しないゲームIDで取得",
			gameID:        testGameID,
			gameVersionID: &gameVersion1ID,
			start:         baseTime,
			end:           baseTime.Add(24 * time.Hour),
			expectedStats: domain.NewGamePlayStats(
				testGameID,
				0,
				0,
				[]*domain.HourlyPlayStats{},
			),
			isErr: false,
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
