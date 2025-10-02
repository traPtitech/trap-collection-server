package gorm2

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
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

	ctx := t.Context()
	db, err := testDB.getDB(ctx)
	require.NoError(t, err)

	now := time.Now()

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
	gamePlayLog1 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	gamePlayLog2 := schema.GamePlayLogTable{
		ID:            uuid.New(),
		EditionID:     edition1.ID,
		GameID:        game1.ID,
		GameVersionID: gameVersion1.ID,
		StartTime:     now.Add(-1 * time.Hour),
		EndTime:       sql.NullTime{Time: now, Valid: true},
		CreatedAt:     now.Add(-1 * time.Hour),
		UpdatedAt:     now,
	}

	require.NoError(t, db.Create(&edition1).Error)
	require.NoError(t, db.Create(&game1).Error)
	require.NoError(t, db.Create(&gameImage1).Error)
	require.NoError(t, db.Create(&gameVideo1).Error)
	require.NoError(t, db.Create(&gameVersion1).Error)
	require.NoError(t, db.Create(&gamePlayLog1).Error)
	require.NoError(t, db.Create(&gamePlayLog2).Error)

	t.Cleanup(func() {
		require.NoError(t, db.WithContext(context.Background()).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GamePlayLogTable{}).Error)
		require.NoError(t, db.WithContext(context.Background()).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVersionTable2{}).Error)
		require.NoError(t, db.WithContext(context.Background()).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVideoTable2{}).Error)
		require.NoError(t, db.WithContext(context.Background()).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameImageTable2{}).Error)
		require.NoError(t, db.WithContext(context.Background()).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error)
		require.NoError(t, db.WithContext(context.Background()).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.EditionTable{}).Error)
	})

	type test struct {
		description     string
		playLogID       values.GamePlayLogID
		expectedPlayLog *domain.GamePlayLog
		expectedErr     error
	}

	testCases := []test{
		{
			description: "正常な場合(EndTimeがNULL)",
			playLogID:   values.GamePlayLogID(gamePlayLog1.ID),
			expectedPlayLog: domain.NewGamePlayLog(
				values.GamePlayLogID(gamePlayLog1.ID),
				values.LauncherVersionID(gamePlayLog1.EditionID),
				values.GameID(gamePlayLog1.GameID),
				values.GameVersionID(gamePlayLog1.GameVersionID),
				gamePlayLog1.StartTime,
				nil,
				gamePlayLog1.CreatedAt,
				gamePlayLog1.UpdatedAt,
			),
		},
		{
			description: "正常な場合(EndTimeが非NULL)",
			playLogID:   values.GamePlayLogID(gamePlayLog2.ID),
			expectedPlayLog: domain.NewGamePlayLog(
				values.GamePlayLogID(gamePlayLog2.ID),
				values.LauncherVersionID(gamePlayLog2.EditionID),
				values.GameID(gamePlayLog2.GameID),
				values.GameVersionID(gamePlayLog2.GameVersionID),
				gamePlayLog2.StartTime,
				&gamePlayLog2.EndTime.Time,
				gamePlayLog2.CreatedAt,
				gamePlayLog2.UpdatedAt,
			),
		},
		{
			description: "存在しない場合",
			playLogID:   values.NewGamePlayLogID(),
			expectedErr: repository.ErrRecordNotFound,
		},
	}

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			playLog, err := gamePlayLogRepository.GetGamePlayLog(ctx, testCase.playLogID)

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
				assert.Nil(t, playLog)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, playLog)

				assert.WithinDuration(t, testCase.expectedPlayLog.GetStartTime(), playLog.GetStartTime(), time.Second)
				if testCase.expectedPlayLog.GetEndTime() != nil {
					assert.NotNil(t, playLog.GetEndTime())
					assert.WithinDuration(t, *testCase.expectedPlayLog.GetEndTime(), *playLog.GetEndTime(), time.Second)
				} else {
					assert.Nil(t, playLog.GetEndTime())
				}

				assert.Equal(t, testCase.expectedPlayLog, playLog, cmpopts.IgnoreFields(domain.GamePlayLog{}, "startTime", "endTime", "createdAt", "updatedAt"))
			}
		})
	}
}

func TestUpdateGamePlayLogEndTime(t *testing.T) {
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

	ctx := context.Background()

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

	// TODO: テストを実装する
	testCases := []test{
		{
			description:   "TODO: add test case",
			gameID:        values.NewGameID(),
			gameVersionID: nil,
			start:         time.Now().Add(-24 * time.Hour),
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
