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
	t.Parallel()

	ctx := context.Background()

	db, err := testDB.getDB(ctx)
	if err != nil {
		t.Fatalf("get db: %+v\n", err)
	}

	gamePlayLogRepository := NewGamePlayLogV2(testDB)

	type test struct {
		description          string
		playLog              *domain.GamePlayLog
		beforeGamePlayLogs   []schema.GamePlayLogTable
		games                []schema.GameTable2
		gameVersions         []schema.GameVersionTable2
		editions             []schema.EditionTable
		expectedGamePlayLogs []schema.GamePlayLogTable
		isErr                bool
		err                  error
	}

	playLogID1 := values.NewGamePlayLogID()
	playLogID2 := values.NewGamePlayLogID()
	playLogID3 := values.NewGamePlayLogID()
	playLogID4 := values.NewGamePlayLogID()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameVersionID3 := values.NewGameVersionID()
	gameVersionID4 := values.NewGameVersionID()

	editionID1 := values.NewLauncherVersionID()
	editionID2 := values.NewLauncherVersionID()
	editionID3 := values.NewLauncherVersionID()
	editionID4 := values.NewLauncherVersionID()

	imageID1 := values.NewGameImageID()
	imageID2 := values.NewGameImageID()
	imageID3 := values.NewGameImageID()
	imageID4 := values.NewGameImageID()
	videoID1 := values.NewGameVideoID()
	videoID2 := values.NewGameVideoID()
	videoID3 := values.NewGameVideoID()
	videoID4 := values.NewGameVideoID()

	now := time.Now()
	startTime1 := now.Add(-1 * time.Hour)
	startTime2 := now.Add(-2 * time.Hour)
	startTime3 := now.Add(-3 * time.Hour)

	var gameVisibilityPublic schema.GameVisibilityTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&schema.GameVisibilityTypeTable{Name: "public"}).
		Find(&gameVisibilityPublic).Error
	if err != nil {
		t.Fatalf("get game visibility: %v\n", err)
	}

	var gameImageType schema.GameImageTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&schema.GameImageTypeTable{Name: "jpeg"}).
		Find(&gameImageType).Error
	if err != nil {
		t.Fatalf("get game image type: %v\n", err)
	}

	var gameVideoType schema.GameVideoTypeTable
	err = db.
		Session(&gorm.Session{}).
		Where(&schema.GameVideoTypeTable{Name: "mp4"}).
		Find(&gameVideoType).Error
	if err != nil {
		t.Fatalf("get game video type: %v\n", err)
	}

	testCases := []test{
		{
			description: "正常にゲームプレイログが作成される",
			playLog: domain.NewGamePlayLog(
				playLogID1,
				editionID1,
				gameID1,
				gameVersionID1,
				startTime1,
				nil,
				now,
				now,
			),
			games: []schema.GameTable2{
				{
					ID:               uuid.UUID(gameID1),
					Name:             "test game 1",
					Description:      "test description 1",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityPublic.ID,
				},
			},
			gameVersions: []schema.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID1),
					GameID:      uuid.UUID(gameID1),
					GameImageID: uuid.UUID(imageID1),
					GameVideoID: uuid.UUID(videoID1),
					Name:        "v1.0.0",
					Description: "test version 1.0.0",
					CreatedAt:   now,
				},
			},
			editions: []schema.EditionTable{
				{
					ID:               uuid.UUID(editionID1),
					Name:             "test edition 1",
					QuestionnaireURL: sql.NullString{String: "", Valid: false},
					CreatedAt:        now,
				},
			},
			expectedGamePlayLogs: []schema.GamePlayLogTable{
				{
					ID:            uuid.UUID(playLogID1),
					EditionID:     uuid.UUID(editionID1),
					GameID:        uuid.UUID(gameID1),
					GameVersionID: uuid.UUID(gameVersionID1),
					StartTime:     startTime1,
					EndTime:       sql.NullTime{},
					CreatedAt:     now,
				},
			},
			isErr: false,
		},
		{
			description: "playLogIDが重複している場合、ErrDuplicatedUniqueKeyが返される",
			playLog: domain.NewGamePlayLog(
				playLogID2,
				editionID2,
				gameID2,
				gameVersionID2,
				startTime2,
				nil,
				now,
				now,
			),
			beforeGamePlayLogs: []schema.GamePlayLogTable{
				{
					ID:            uuid.UUID(playLogID2),
					EditionID:     uuid.UUID(editionID2),
					GameID:        uuid.UUID(gameID2),
					GameVersionID: uuid.UUID(gameVersionID2),
					StartTime:     startTime2,
					EndTime:       sql.NullTime{},
					CreatedAt:     now,
				},
			},
			games: []schema.GameTable2{
				{
					ID:               uuid.UUID(gameID2),
					Name:             "test game 2",
					Description:      "test description 2",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityPublic.ID,
				},
			},
			gameVersions: []schema.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID2),
					GameID:      uuid.UUID(gameID2),
					GameImageID: uuid.UUID(imageID2),
					GameVideoID: uuid.UUID(videoID2),
					Name:        "v1.0.0",
					Description: "test version 1.0.0",
					CreatedAt:   now,
				},
			},
			editions: []schema.EditionTable{
				{
					ID:               uuid.UUID(editionID2),
					Name:             "test edition 2",
					QuestionnaireURL: sql.NullString{String: "", Valid: false},
					CreatedAt:        now,
				},
			},
			expectedGamePlayLogs: []schema.GamePlayLogTable{
				{
					ID:            uuid.UUID(playLogID2),
					EditionID:     uuid.UUID(editionID2),
					GameID:        uuid.UUID(gameID2),
					GameVersionID: uuid.UUID(gameVersionID2),
					StartTime:     startTime2,
					EndTime:       sql.NullTime{},
					CreatedAt:     now,
				},
			},
			isErr: true,
			err:   repository.ErrDuplicatedUniqueKey,
		},
		{
			description: "既存のログが存在していても新しいログを正常に作成できる",
			playLog: domain.NewGamePlayLog(
				playLogID3,
				editionID4,
				gameID4,
				gameVersionID4,
				startTime3,
				nil,
				now,
				now,
			),
			beforeGamePlayLogs: []schema.GamePlayLogTable{
				{
					ID:            uuid.UUID(playLogID4),
					EditionID:     uuid.UUID(editionID3),
					GameID:        uuid.UUID(gameID3),
					GameVersionID: uuid.UUID(gameVersionID3),
					StartTime:     startTime1,
					EndTime:       sql.NullTime{},
					CreatedAt:     now,
				},
			},
			games: []schema.GameTable2{
				{
					ID:               uuid.UUID(gameID3),
					Name:             "test game 3 existing",
					Description:      "test description 3 existing",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityPublic.ID,
				},
				{
					ID:               uuid.UUID(gameID4),
					Name:             "test game 4 new",
					Description:      "test description 4 new",
					CreatedAt:        now,
					VisibilityTypeID: gameVisibilityPublic.ID,
				},
			},
			gameVersions: []schema.GameVersionTable2{
				{
					ID:          uuid.UUID(gameVersionID3),
					GameID:      uuid.UUID(gameID3),
					GameImageID: uuid.UUID(imageID3),
					GameVideoID: uuid.UUID(videoID3),
					Name:        "v1.1.0",
					Description: "test version 1.1.0",
					CreatedAt:   now,
				},
				{
					ID:          uuid.UUID(gameVersionID4),
					GameID:      uuid.UUID(gameID4),
					GameImageID: uuid.UUID(imageID4),
					GameVideoID: uuid.UUID(videoID4),
					Name:        "v1.2.0",
					Description: "test version 1.2.0",
					CreatedAt:   now,
				},
			},
			editions: []schema.EditionTable{
				{
					ID:               uuid.UUID(editionID3),
					Name:             "test edition 3",
					QuestionnaireURL: sql.NullString{String: "", Valid: false},
					CreatedAt:        now,
				},
				{
					ID:               uuid.UUID(editionID4),
					Name:             "test edition 4",
					QuestionnaireURL: sql.NullString{String: "", Valid: false},
					CreatedAt:        now,
				},
			},
			expectedGamePlayLogs: []schema.GamePlayLogTable{
				{
					ID:            uuid.UUID(playLogID3),
					EditionID:     uuid.UUID(editionID4),
					GameID:        uuid.UUID(gameID4),
					GameVersionID: uuid.UUID(gameVersionID4),
					StartTime:     startTime3,
					EndTime:       sql.NullTime{},
					CreatedAt:     now,
				},
			},
			isErr: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			// 並列実行対応のため，テストケースごとにレコード単位で削除
			t.Cleanup(func() {
				if testCase.playLog != nil {
					db.Session(&gorm.Session{}).Unscoped().
						Where("id = ?", uuid.UUID(testCase.playLog.GetID())).
						Delete(&schema.GamePlayLogTable{})
				}
				for _, log := range testCase.beforeGamePlayLogs {
					db.Session(&gorm.Session{}).Unscoped().
						Where("id = ?", log.ID).
						Delete(&schema.GamePlayLogTable{})
				}
				for _, version := range testCase.gameVersions {
					db.Session(&gorm.Session{}).Unscoped().
						Where("id = ?", version.ID).
						Delete(&schema.GameVersionTable2{})
					db.Session(&gorm.Session{}).Unscoped().
						Where("id = ?", version.GameImageID).
						Delete(&schema.GameImageTable2{})
					db.Session(&gorm.Session{}).Unscoped().
						Where("id = ?", version.GameVideoID).
						Delete(&schema.GameVideoTable2{})
				}
				for _, game := range testCase.games {
					db.Session(&gorm.Session{}).Unscoped().
						Where("id = ?", game.ID).
						Delete(&schema.GameTable2{})
				}
				for _, edition := range testCase.editions {
					db.Session(&gorm.Session{}).Unscoped().
						Where("id = ?", edition.ID).
						Delete(&schema.EditionTable{})
				}
			})

			if len(testCase.games) != 0 {
				err := db.
					Session(&gorm.Session{}).
					Create(&testCase.games).Error
				if err != nil {
					t.Fatalf("create games: %+v\n", err)
				}
			}

			if len(testCase.editions) != 0 {
				err := db.
					Session(&gorm.Session{}).
					Create(&testCase.editions).Error
				if err != nil {
					t.Fatalf("create editions: %+v\n", err)
				}
			}

			if len(testCase.gameVersions) != 0 {
				for _, version := range testCase.gameVersions {
					image := schema.GameImageTable2{
						ID:          version.GameImageID,
						GameID:      version.GameID,
						ImageTypeID: gameImageType.ID,
						CreatedAt:   now,
					}
					err := db.Session(&gorm.Session{}).Create(&image).Error
					if err != nil {
						t.Fatalf("create game image: %+v\n", err)
					}

					video := schema.GameVideoTable2{
						ID:          version.GameVideoID,
						GameID:      version.GameID,
						VideoTypeID: gameVideoType.ID,
						CreatedAt:   now,
					}
					err = db.Session(&gorm.Session{}).Create(&video).Error
					if err != nil {
						t.Fatalf("create game video: %+v\n", err)
					}
				}

				err := db.
					Session(&gorm.Session{}).
					Create(&testCase.gameVersions).Error
				if err != nil {
					t.Fatalf("create game versions: %+v\n", err)
				}
			}

			if len(testCase.beforeGamePlayLogs) != 0 {
				err := db.
					Session(&gorm.Session{}).
					Create(&testCase.beforeGamePlayLogs).Error
				if err != nil {
					t.Fatalf("create before game play logs: %+v\n", err)
				}
			}

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

			for _, expectedLog := range testCase.expectedGamePlayLogs {
				var actualLog schema.GamePlayLogTable
				err = db.
					Session(&gorm.Session{}).
					Where("id = ?", expectedLog.ID).
					First(&actualLog).Error

				assert.NoErrorf(t, err, "expected log with ID %v not found", expectedLog.ID)
				assert.Equal(t, expectedLog.EditionID, actualLog.EditionID)
				assert.Equal(t, expectedLog.GameID, actualLog.GameID)
				assert.Equal(t, expectedLog.GameVersionID, actualLog.GameVersionID)
				assert.WithinDuration(t, expectedLog.StartTime, actualLog.StartTime, time.Second)
				assert.Equal(t, expectedLog.EndTime.Valid, actualLog.EndTime.Valid)
				if expectedLog.EndTime.Valid {
					assert.WithinDuration(t, expectedLog.EndTime.Time, actualLog.EndTime.Time, time.Second)
				}
				assert.WithinDuration(t, expectedLog.CreatedAt, actualLog.CreatedAt, time.Second)
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
		ctx := context.Background()
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GamePlayLogTable{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVersionTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameVideoTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameImageTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.GameTable2{}).Error)
		require.NoError(t, db.WithContext(ctx).Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&schema.EditionTable{}).Error)
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
				assert.Equal(t, testCase.expectedPlayLog, playLog)
			}
		})
	}
}

func TestUpdateGamePlayLogEndTime(t *testing.T) {
	t.Skip("この機能はまだ実装されていないため、テストをスキップします。")
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
	t.Skip("この機能はまだ実装されていないため、テストをスキップします。")
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
	t.Skip("この機能はまだ実装されていないため、テストをスキップします。")
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
