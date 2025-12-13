package v2

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	mockRepository "github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
	"go.uber.org/mock/gomock"
)

func TestCreatePlayLog(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type test struct {
		description   string
		editionID     values.EditionID
		gameID        values.GameID
		gameVersionID values.GameVersionID
		startTime     time.Time

		executeGetEdition bool
		getEditionResult  *domain.Edition
		getEditionErr     error

		executeGetGame bool
		getGameResult  *domain.Game
		getGameErr     error

		executeGetGameVersionByID bool
		getGameVersionByIDResult  *repository.GameVersionInfoWithGameID
		getGameVersionByIDErr     error

		executeCreateGamePlayLog bool
		createGamePlayLogErr     error

		isErr bool
		err   error
	}

	now := time.Now()
	editionID := values.NewEditionID()
	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()

	questionnaireURL, _ := url.Parse("https://example.com")
	edition := domain.NewEditionWithQuestionnaire(
		editionID,
		values.NewEditionName("v1.0.0"),
		values.NewEditionQuestionnaireURL(questionnaireURL),
		now,
	)

	game := domain.NewGame(
		gameID,
		values.NewGameName("Test Game"),
		values.NewGameDescription("Test Description"),
		values.GameVisibilityTypePublic,
		now,
	)

	gameVersion := domain.NewGameVersion(
		gameVersionID,
		values.NewGameVersionName("v1.0.0"),
		values.NewGameVersionDescription("Test Version"),
		now,
	)

	testCases := []test{
		{
			description:               "正常にプレイログが作成される",
			editionID:                 editionID,
			gameID:                    gameID,
			gameVersionID:             gameVersionID,
			startTime:                 now,
			executeGetEdition:         true,
			getEditionResult:          edition,
			executeGetGame:            true,
			getGameResult:             game,
			executeGetGameVersionByID: true,
			getGameVersionByIDResult:  &repository.GameVersionInfoWithGameID{GameVersion: gameVersion, GameID: gameID},
			executeCreateGamePlayLog:  true,
			isErr:                     false,
		},
		{
			description:       "GetEditionがErrRecordNotFoundなのでErrInvalidEdition",
			editionID:         values.NewEditionID(),
			gameID:            gameID,
			gameVersionID:     gameVersionID,
			startTime:         now,
			executeGetEdition: true,
			getEditionErr:     repository.ErrRecordNotFound,
			isErr:             true,
			err:               service.ErrInvalidEdition,
		},
		{
			description:       "GetEditionがエラーなのでエラー",
			editionID:         editionID,
			gameID:            gameID,
			gameVersionID:     gameVersionID,
			startTime:         now,
			executeGetEdition: true,
			getEditionErr:     assert.AnError,
			isErr:             true,
			err:               assert.AnError,
		},
		{
			description:       "GetGameがErrRecordNotFoundなのでErrInvalidGame",
			editionID:         editionID,
			gameID:            values.NewGameID(),
			gameVersionID:     gameVersionID,
			startTime:         now,
			executeGetEdition: true,
			getEditionResult:  edition,
			executeGetGame:    true,
			getGameErr:        repository.ErrRecordNotFound,
			isErr:             true,
			err:               service.ErrInvalidGame,
		},
		{
			description:       "GetGameがエラーなのでエラー",
			editionID:         editionID,
			gameID:            gameID,
			gameVersionID:     gameVersionID,
			startTime:         now,
			executeGetEdition: true,
			getEditionResult:  edition,
			executeGetGame:    true,
			getGameErr:        assert.AnError,
			isErr:             true,
			err:               assert.AnError,
		},
		{
			description:               "GetGameVersionByIDがErrRecordNotFoundなのでErrInvalidGameVersion",
			editionID:                 editionID,
			gameID:                    gameID,
			gameVersionID:             values.NewGameVersionID(),
			startTime:                 now,
			executeGetEdition:         true,
			getEditionResult:          edition,
			executeGetGame:            true,
			getGameResult:             game,
			executeGetGameVersionByID: true,
			getGameVersionByIDErr:     repository.ErrRecordNotFound,
			isErr:                     true,
			err:                       service.ErrInvalidGameVersion,
		},
		{
			description:               "GetGameVersionByIDがエラーなのでエラー",
			editionID:                 editionID,
			gameID:                    gameID,
			gameVersionID:             gameVersionID,
			startTime:                 now,
			executeGetEdition:         true,
			getEditionResult:          edition,
			executeGetGame:            true,
			getGameResult:             game,
			executeGetGameVersionByID: true,
			getGameVersionByIDErr:     assert.AnError,
			isErr:                     true,
			err:                       assert.AnError,
		},
		{
			description:               "CreateGamePlayLogがエラーなのでエラー",
			editionID:                 editionID,
			gameID:                    gameID,
			gameVersionID:             gameVersionID,
			startTime:                 now,
			executeGetEdition:         true,
			getEditionResult:          edition,
			executeGetGame:            true,
			getGameResult:             game,
			executeGetGameVersionByID: true,
			getGameVersionByIDResult:  &repository.GameVersionInfoWithGameID{GameVersion: gameVersion, GameID: gameID},
			executeCreateGamePlayLog:  true,
			createGamePlayLogErr:      assert.AnError,
			isErr:                     true,
			err:                       assert.AnError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDB := mockRepository.NewMockDB(ctrl)
			mockGamePlayLogRepository := mockRepository.NewMockGamePlayLogV2(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

			gamePlayLogService := NewGamePlayLog(
				mockDB,
				mockGamePlayLogRepository,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
			)

			if testCase.executeGetEdition {
				mockEditionRepository.
					EXPECT().
					GetEdition(ctx, testCase.editionID, repository.LockTypeNone).
					Return(testCase.getEditionResult, testCase.getEditionErr)
			}

			if testCase.executeGetGame {
				mockGameRepository.
					EXPECT().
					GetGame(ctx, testCase.gameID, repository.LockTypeNone).
					Return(testCase.getGameResult, testCase.getGameErr)
			}

			if testCase.executeGetGameVersionByID {
				mockGameVersionRepository.
					EXPECT().
					GetGameVersionByID(ctx, testCase.gameVersionID, repository.LockTypeNone).
					Return(testCase.getGameVersionByIDResult, testCase.getGameVersionByIDErr)
			}

			if testCase.executeCreateGamePlayLog {
				mockGamePlayLogRepository.
					EXPECT().
					CreateGamePlayLog(ctx, gomock.Cond(func(playLog *domain.GamePlayLog) bool {
						if playLog.GetEditionID() != testCase.editionID {
							t.Errorf("EditionID: expected %v, got %v", testCase.editionID, playLog.GetEditionID())
							return false
						}
						if playLog.GetGameID() != testCase.gameID {
							t.Errorf("GameID: expected %v, got %v", testCase.gameID, playLog.GetGameID())
							return false
						}
						if playLog.GetGameVersionID() != testCase.gameVersionID {
							t.Errorf("GameVersionID: expected %v, got %v", testCase.gameVersionID, playLog.GetGameVersionID())
							return false
						}
						if !playLog.GetStartTime().Equal(testCase.startTime) {
							t.Errorf("StartTime: expected %v, got %v", testCase.startTime, playLog.GetStartTime())
							return false
						}
						if playLog.GetEndTime() != nil {
							t.Errorf("EndTime: expected nil, got %v", playLog.GetEndTime())
							return false
						}
						if playLog.GetID() == values.GamePlayLogID(uuid.Nil) {
							t.Error("ID should not be nil")
							return false
						}
						if time.Since(playLog.GetCreatedAt()) > time.Second {
							t.Errorf("CreatedAt should be recent: %v", playLog.GetCreatedAt())
							return false
						}
						if time.Since(playLog.GetUpdatedAt()) > time.Second {
							t.Errorf("UpdatedAt should be recent: %v", playLog.GetUpdatedAt())
							return false
						}
						return true
					})).
					Return(testCase.createGamePlayLogErr)
			}

			playLog, err := gamePlayLogService.CreatePlayLog(
				ctx,
				testCase.editionID,
				testCase.gameID,
				testCase.gameVersionID,
				testCase.startTime,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, playLog)
				assert.Equal(t, testCase.editionID, playLog.GetEditionID())
				assert.Equal(t, testCase.gameID, playLog.GetGameID())
				assert.Equal(t, testCase.gameVersionID, playLog.GetGameVersionID())
				assert.Equal(t, testCase.startTime, playLog.GetStartTime())
				assert.Nil(t, playLog.GetEndTime())
			}
		})
	}
}

func TestUpdatePlayLogEndTime(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type test struct {
		description string
		playLogID   values.GamePlayLogID
		endTime     time.Time

		executeGetGamePlayLog bool
		getGamePlayLogResult  *domain.GamePlayLog
		getGamePlayLogErr     error

		executeUpdateGamePlayLogEndTime bool
		updateGamePlayLogEndTimeErr     error

		isErr bool
		err   error
	}

	now := time.Now()
	playLogID := values.NewGamePlayLogID()
	editionID := values.NewEditionID()
	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()

	// プレイ中のログ（EndTimeがnil）
	activePlayLog := domain.NewGamePlayLog(
		playLogID,
		editionID,
		gameID,
		gameVersionID,
		now.Add(-time.Hour),
		nil,
		now.Add(-time.Hour),
		now.Add(-time.Hour),
	)

	// 異なるeditionIDとgameIDを持つプレイログ
	mismatchedPlayLog := domain.NewGamePlayLog(
		playLogID,
		values.NewEditionID(),
		values.NewGameID(),
		gameVersionID,
		now.Add(-time.Hour),
		nil,
		now.Add(-time.Hour),
		now.Add(-time.Hour),
	)

	testCases := []test{
		{
			description:                     "正常にプレイログが終了される",
			playLogID:                       playLogID,
			endTime:                         now,
			executeGetGamePlayLog:           true,
			getGamePlayLogResult:            activePlayLog,
			executeUpdateGamePlayLogEndTime: true,
			isErr:                           false,
		},
		{
			description:           "GetGamePlayLogがErrRecordNotFoundなのでErrInvalidPlayLogID",
			playLogID:             values.NewGamePlayLogID(),
			endTime:               now,
			executeGetGamePlayLog: true,
			getGamePlayLogErr:     repository.ErrRecordNotFound,
			isErr:                 true,
			err:                   service.ErrInvalidPlayLogID,
		},
		{
			description:           "GetGamePlayLogがエラーなのでエラー",
			playLogID:             playLogID,
			endTime:               now,
			executeGetGamePlayLog: true,
			getGamePlayLogErr:     assert.AnError,
			isErr:                 true,
			err:                   assert.AnError,
		},
		{
			description:           "終了時刻が開始時刻より前なのでErrInvalidEndTime",
			playLogID:             playLogID,
			endTime:               activePlayLog.GetStartTime().Add(-time.Hour), // StartTimeより前
			executeGetGamePlayLog: true,
			getGamePlayLogResult:  activePlayLog,
			isErr:                 true,
			err:                   service.ErrInvalidEndTime,
		},
		{
			description:                     "UpdateGamePlayLogEndTimeがエラーなのでエラー",
			playLogID:                       playLogID,
			endTime:                         now,
			executeGetGamePlayLog:           true,
			getGamePlayLogResult:            activePlayLog,
			executeUpdateGamePlayLogEndTime: true,
			updateGamePlayLogEndTimeErr:     assert.AnError,
			isErr:                           true,
			err:                             assert.AnError,
		},
		{
			description:                     "開始時刻と同じ時刻でも正常に終了できる",
			playLogID:                       playLogID,
			endTime:                         activePlayLog.GetStartTime(),
			executeGetGamePlayLog:           true,
			getGamePlayLogResult:            activePlayLog,
			executeUpdateGamePlayLogEndTime: true,
			isErr:                           false,
		},
		{
			description:                     "プレイログがeditionIDとgameIDのペアに対応しない場合はErrInvalidPlayLogEditionGamePairエラー",
			playLogID:                       playLogID,
			endTime:                         now,
			executeGetGamePlayLog:           true,
			getGamePlayLogResult:            mismatchedPlayLog,
			executeUpdateGamePlayLogEndTime: false,
			isErr:                           true,
			err:                             service.ErrInvalidPlayLogEditionGamePair,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDB := mockRepository.NewMockDB(ctrl)
			mockGamePlayLogRepository := mockRepository.NewMockGamePlayLogV2(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

			gamePlayLogService := NewGamePlayLog(
				mockDB,
				mockGamePlayLogRepository,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
			)

			if testCase.executeGetGamePlayLog {
				mockGamePlayLogRepository.
					EXPECT().
					GetGamePlayLog(ctx, testCase.playLogID).
					Return(testCase.getGamePlayLogResult, testCase.getGamePlayLogErr)
			}

			if testCase.executeUpdateGamePlayLogEndTime {
				mockGamePlayLogRepository.
					EXPECT().
					UpdateGamePlayLogEndTime(ctx, testCase.playLogID, testCase.endTime).
					Return(testCase.updateGamePlayLogEndTimeErr)
			}

			err := gamePlayLogService.UpdatePlayLogEndTime(
				ctx,
				editionID,
				gameID,
				testCase.playLogID,
				testCase.endTime,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
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

	type test struct {
		description   string
		gameID        values.GameID
		gameVersionID *values.GameVersionID
		start         time.Time
		end           time.Time

		executeGetGame bool
		getGameResult  *domain.Game
		getGameErr     error

		executeGetGameVersionByID bool
		getGameVersionByIDResult  *repository.GameVersionInfoWithGameID
		getGameVersionByIDErr     error

		executeGetGamePlayStats bool
		getGamePlayStatsResult  *domain.GamePlayStats
		getGamePlayStatsErr     error

		isErr bool
		err   error
	}

	now := time.Now()
	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()

	game := domain.NewGame(
		gameID,
		values.NewGameName("Test Game"),
		values.NewGameDescription("Test Description"),
		values.GameVisibilityTypePublic,
		now.Add(-24*time.Hour),
	)

	gameVersion := domain.NewGameVersion(
		gameVersionID,
		values.NewGameVersionName("v1.0.0"),
		values.NewGameVersionDescription("Test Version"),
		now.Add(-24*time.Hour),
	)

	sampleStats := domain.NewGamePlayStats(
		gameID,
		10,
		3600*time.Second,
		[]*domain.HourlyPlayStats{
			domain.NewHourlyPlayStats(
				now.Add(-2*time.Hour).Truncate(time.Hour),
				5,
				1800*time.Second,
			),
			domain.NewHourlyPlayStats(
				now.Add(-1*time.Hour).Truncate(time.Hour),
				5,
				1800*time.Second,
			),
		},
	)

	testCases := []test{
		{
			description:             "gameVersionIDなしで正常に統計が取得される",
			gameID:                  gameID,
			gameVersionID:           nil,
			start:                   now.Add(-24 * time.Hour),
			end:                     now,
			executeGetGame:          true,
			getGameResult:           game,
			executeGetGamePlayStats: true,
			getGamePlayStatsResult:  sampleStats,
			isErr:                   false,
		},
		{
			description:               "gameVersionIDありで正常に統計が取得される",
			gameID:                    gameID,
			gameVersionID:             &gameVersionID,
			start:                     now.Add(-24 * time.Hour),
			end:                       now,
			executeGetGame:            true,
			getGameResult:             game,
			executeGetGameVersionByID: true,
			getGameVersionByIDResult:  &repository.GameVersionInfoWithGameID{GameVersion: gameVersion, GameID: gameID},
			executeGetGamePlayStats:   true,
			getGamePlayStatsResult:    sampleStats,
			isErr:                     false,
		},
		{
			description:    "GetGameがErrRecordNotFoundなのでErrInvalidGame",
			gameID:         values.NewGameID(),
			gameVersionID:  nil,
			start:          now.Add(-24 * time.Hour),
			end:            now,
			executeGetGame: true,
			getGameErr:     repository.ErrRecordNotFound,
			isErr:          true,
			err:            service.ErrInvalidGame,
		},
		{
			description:    "GetGameがエラーなのでエラー",
			gameID:         gameID,
			gameVersionID:  nil,
			start:          now.Add(-24 * time.Hour),
			end:            now,
			executeGetGame: true,
			getGameErr:     assert.AnError,
			isErr:          true,
			err:            assert.AnError,
		},
		{
			description:               "GetGameVersionByIDがErrRecordNotFoundなのでErrInvalidGameVersion",
			gameID:                    gameID,
			gameVersionID:             &gameVersionID,
			start:                     now.Add(-24 * time.Hour),
			end:                       now,
			executeGetGame:            true,
			getGameResult:             game,
			executeGetGameVersionByID: true,
			getGameVersionByIDErr:     repository.ErrRecordNotFound,
			isErr:                     true,
			err:                       service.ErrInvalidGameVersion,
		},
		{
			description:               "GetGameVersionByIDがエラーなのでエラー",
			gameID:                    gameID,
			gameVersionID:             &gameVersionID,
			start:                     now.Add(-24 * time.Hour),
			end:                       now,
			executeGetGame:            true,
			getGameResult:             game,
			executeGetGameVersionByID: true,
			getGameVersionByIDErr:     assert.AnError,
			isErr:                     true,
			err:                       assert.AnError,
		},
		{
			description:             "GetGamePlayStatsがエラーなのでエラー",
			gameID:                  gameID,
			gameVersionID:           nil,
			start:                   now.Add(-24 * time.Hour),
			end:                     now,
			executeGetGame:          true,
			getGameResult:           game,
			executeGetGamePlayStats: true,
			getGamePlayStatsErr:     assert.AnError,
			isErr:                   true,
			err:                     assert.AnError,
		},
		{
			description:             "開始時刻と終了時刻が同じでも正常に取得できる",
			gameID:                  gameID,
			gameVersionID:           nil,
			start:                   now,
			end:                     now,
			executeGetGame:          true,
			getGameResult:           game,
			executeGetGamePlayStats: true,
			getGamePlayStatsResult: domain.NewGamePlayStats(
				gameID,
				0,
				0,
				[]*domain.HourlyPlayStats{},
			),
			isErr: false,
		},
		{
			description:   "終了時刻が開始時刻より前なのでErrInvalidTimeRange",
			gameID:        gameID,
			gameVersionID: nil,
			start:         now,
			end:           now.Add(-1 * time.Hour),
			isErr:         true,
			err:           service.ErrInvalidTimeRange,
		},
		{
			description:   "期間が10年を超えるのでErrTimePeriodTooLong",
			gameID:        gameID,
			gameVersionID: nil,
			start:         now.AddDate(-10, 0, -1),
			end:           now,
			isErr:         true,
			err:           service.ErrTimePeriodTooLong,
		},
		{
			description:             "期間がちょうど10年なら正常に取得できる",
			gameID:                  gameID,
			gameVersionID:           nil,
			start:                   now.AddDate(-10, 0, 0),
			end:                     now,
			executeGetGame:          true,
			getGameResult:           game,
			executeGetGamePlayStats: true,
			getGamePlayStatsResult:  sampleStats,
			isErr:                   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDB := mockRepository.NewMockDB(ctrl)
			mockGamePlayLogRepository := mockRepository.NewMockGamePlayLogV2(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

			gamePlayLogService := NewGamePlayLog(
				mockDB,
				mockGamePlayLogRepository,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
			)

			if testCase.executeGetGame {
				mockGameRepository.
					EXPECT().
					GetGame(ctx, testCase.gameID, repository.LockTypeNone).
					Return(testCase.getGameResult, testCase.getGameErr)
			}

			if testCase.executeGetGameVersionByID {
				mockGameVersionRepository.
					EXPECT().
					GetGameVersionByID(ctx, *testCase.gameVersionID, repository.LockTypeNone).
					Return(testCase.getGameVersionByIDResult, testCase.getGameVersionByIDErr)
			}

			if testCase.executeGetGamePlayStats {
				mockGamePlayLogRepository.
					EXPECT().
					GetGamePlayStats(ctx, testCase.gameID, testCase.gameVersionID, testCase.start, testCase.end).
					Return(testCase.getGamePlayStatsResult, testCase.getGamePlayStatsErr)
			}

			stats, err := gamePlayLogService.GetGamePlayStats(
				ctx,
				testCase.gameID,
				testCase.gameVersionID,
				testCase.start,
				testCase.end,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				assert.Equal(t, testCase.getGamePlayStatsResult.GetGameID(), stats.GetGameID())
				assert.Equal(t, testCase.getGamePlayStatsResult.GetTotalPlayCount(), stats.GetTotalPlayCount())
				assert.Equal(t, testCase.getGamePlayStatsResult.GetTotalPlayTime(), stats.GetTotalPlayTime())
				assert.Equal(t, len(testCase.getGamePlayStatsResult.GetHourlyStats()), len(stats.GetHourlyStats()))
			}
		})
	}
}

func TestGetEditionPlayStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	type test struct {
		description string
		editionID   values.EditionID
		start       time.Time
		end         time.Time

		executeGetEdition bool
		getEditionResult  *domain.Edition
		getEditionErr     error

		executeGetEditionPlayStats bool
		getEditionPlayStatsResult  *domain.EditionPlayStats
		getEditionPlayStatsErr     error

		isErr bool
		err   error
	}

	now := time.Now()
	editionID := values.NewEditionID()
	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	questionnaireURL, _ := url.Parse("https://example.com")
	edition := domain.NewEditionWithQuestionnaire(
		editionID,
		values.NewEditionName("v1.0.0"),
		values.NewEditionQuestionnaireURL(questionnaireURL),
		now,
	)

	sampleEditionStats := domain.NewEditionPlayStats(
		editionID,
		values.NewEditionName("v1.0.0"),
		15,
		5400*time.Second,
		[]*domain.GamePlayStatsInEdition{
			domain.NewGamePlayStatsInEdition(
				gameID1,
				8,
				3200*time.Second,
			),
			domain.NewGamePlayStatsInEdition(
				gameID2,
				7,
				2200*time.Second,
			),
		},
		[]*domain.HourlyPlayStats{
			domain.NewHourlyPlayStats(
				now.Add(-3*time.Hour).Truncate(time.Hour),
				5,
				1800*time.Second,
			),
			domain.NewHourlyPlayStats(
				now.Add(-2*time.Hour).Truncate(time.Hour),
				5,
				1800*time.Second,
			),
			domain.NewHourlyPlayStats(
				now.Add(-1*time.Hour).Truncate(time.Hour),
				5,
				1800*time.Second,
			),
		},
	)

	testCases := []test{
		{
			description:                "正常にエディション統計が取得される",
			editionID:                  editionID,
			start:                      now.Add(-24 * time.Hour),
			end:                        now,
			executeGetEdition:          true,
			getEditionResult:           edition,
			executeGetEditionPlayStats: true,
			getEditionPlayStatsResult:  sampleEditionStats,
			isErr:                      false,
		},
		{
			description:       "GetEditionがErrRecordNotFoundなのでErrInvalidEdition",
			editionID:         values.NewEditionID(),
			start:             now.Add(-24 * time.Hour),
			end:               now,
			executeGetEdition: true,
			getEditionErr:     repository.ErrRecordNotFound,
			isErr:             true,
			err:               service.ErrInvalidEdition,
		},
		{
			description:       "GetEditionがエラーなのでエラー",
			editionID:         editionID,
			start:             now.Add(-24 * time.Hour),
			end:               now,
			executeGetEdition: true,
			getEditionErr:     assert.AnError,
			isErr:             true,
			err:               assert.AnError,
		},
		{
			description:                "GetEditionPlayStatsがエラーなのでエラー",
			editionID:                  editionID,
			start:                      now.Add(-24 * time.Hour),
			end:                        now,
			executeGetEdition:          true,
			getEditionResult:           edition,
			executeGetEditionPlayStats: true,
			getEditionPlayStatsErr:     assert.AnError,
			isErr:                      true,
			err:                        assert.AnError,
		},
		{
			description:                "開始時刻と終了時刻が同じでも正常に取得できる",
			editionID:                  editionID,
			start:                      now,
			end:                        now,
			executeGetEdition:          true,
			getEditionResult:           edition,
			executeGetEditionPlayStats: true,
			getEditionPlayStatsResult: domain.NewEditionPlayStats(
				editionID,
				values.NewEditionName("v1.0.0"),
				0,
				0,
				[]*domain.GamePlayStatsInEdition{},
				[]*domain.HourlyPlayStats{},
			),
			isErr: false,
		},
		{
			description:       "終了時刻が開始時刻より前なのでErrInvalidTimeRange",
			editionID:         editionID,
			start:             now,
			end:               now.Add(-1 * time.Hour),
			executeGetEdition: false,
			isErr:             true,
			err:               service.ErrInvalidTimeRange,
		},
		{
			description:       "期間が10年を超えるのでErrTimePeriodTooLong",
			editionID:         editionID,
			start:             now.AddDate(-10, 0, -1),
			end:               now,
			executeGetEdition: false,
			isErr:             true,
			err:               service.ErrTimePeriodTooLong,
		},
		{
			description:                "期間がちょうど10年なら正常に取得できる",
			editionID:                  editionID,
			start:                      now.AddDate(-10, 0, 0),
			end:                        now,
			executeGetEdition:          true,
			getEditionResult:           edition,
			executeGetEditionPlayStats: true,
			getEditionPlayStatsResult:  sampleEditionStats,
			isErr:                      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDB := mockRepository.NewMockDB(ctrl)
			mockGamePlayLogRepository := mockRepository.NewMockGamePlayLogV2(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

			gamePlayLogService := NewGamePlayLog(
				mockDB,
				mockGamePlayLogRepository,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
			)

			if testCase.executeGetEdition {
				mockEditionRepository.
					EXPECT().
					GetEdition(ctx, testCase.editionID, repository.LockTypeNone).
					Return(testCase.getEditionResult, testCase.getEditionErr)
			}

			if testCase.executeGetEditionPlayStats {
				mockGamePlayLogRepository.
					EXPECT().
					GetEditionPlayStats(ctx, testCase.editionID, testCase.start, testCase.end).
					Return(testCase.getEditionPlayStatsResult, testCase.getEditionPlayStatsErr)
			}

			stats, err := gamePlayLogService.GetEditionPlayStats(
				ctx,
				testCase.editionID,
				testCase.start,
				testCase.end,
			)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				assert.Equal(t, testCase.editionID, stats.GetEditionID())
				assert.Equal(t, testCase.getEditionPlayStatsResult.GetEditionName(), stats.GetEditionName())
				assert.Equal(t, testCase.getEditionPlayStatsResult.GetTotalPlayCount(), stats.GetTotalPlayCount())
				assert.Equal(t, testCase.getEditionPlayStatsResult.GetTotalPlayTime(), stats.GetTotalPlayTime())
				assert.Equal(t, len(testCase.getEditionPlayStatsResult.GetGameStats()), len(stats.GetGameStats()))
				assert.Equal(t, len(testCase.getEditionPlayStatsResult.GetHourlyStats()), len(stats.GetHourlyStats()))
			}
		})
	}
}

func TestDeleteGamePlayLog(t *testing.T) {
	t.Parallel()

	editionID := values.NewEditionID()
	gameID := values.NewGameID()
	playLogID := values.NewGamePlayLogID()
	playLog := domain.NewGamePlayLog(
		playLogID,
		editionID,
		gameID,
		values.NewGameVersionID(),
		time.Now().Add(-2*time.Hour),
		nil,
		time.Now().Add(-2*time.Hour),
		time.Now().Add(-2*time.Hour),
	)

	testCases := map[string]struct {
		editionID                values.EditionID
		gameID                   values.GameID
		playLogID                values.GamePlayLogID
		GetGamePlayLogErr        error
		executeDeleteGamePlayLog bool
		deleteGamePlayLogErr     error
		playLog                  *domain.GamePlayLog
		err                      error
	}{
		"GetGamePlayLogがErrRecordNotFoundなのでErrInvalidPlayLogID": {
			editionID:         values.NewEditionID(),
			gameID:            values.NewGameID(),
			playLogID:         values.NewGamePlayLogID(),
			GetGamePlayLogErr: repository.ErrRecordNotFound,
			err:               service.ErrInvalidPlayLogID,
		},
		"GetGamePlayLogがエラーなのでエラー": {
			editionID:         values.NewEditionID(),
			gameID:            values.NewGameID(),
			playLogID:         values.NewGamePlayLogID(),
			GetGamePlayLogErr: assert.AnError,
			err:               assert.AnError,
		},
		"editionIDとgameIDのペアが異なるのでErrInvalidPlayLogID": {
			editionID: values.NewEditionID(),
			gameID:    values.NewGameID(),
			playLogID: playLogID,
			playLog:   playLog,
			err:       service.ErrInvalidPlayLogID,
		},
		"DeleteGamePlayLogがErrNoRecordDeletedなのでErrInvalidPlayLogID": {
			editionID:                editionID,
			gameID:                   gameID,
			playLogID:                playLogID,
			playLog:                  playLog,
			executeDeleteGamePlayLog: true,
			deleteGamePlayLogErr:     repository.ErrNoRecordDeleted,
			err:                      service.ErrInvalidPlayLogID,
		},
		"DeleteGamePlayLogがエラーなのでエラー": {
			editionID:                editionID,
			gameID:                   gameID,
			playLogID:                playLogID,
			playLog:                  playLog,
			executeDeleteGamePlayLog: true,
			deleteGamePlayLogErr:     assert.AnError,
			err:                      assert.AnError,
		},
		"正常にプレイログが削除される": {
			editionID:                editionID,
			gameID:                   gameID,
			playLogID:                playLogID,
			playLog:                  playLog,
			executeDeleteGamePlayLog: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDB := mockRepository.NewMockDB(ctrl)
			mockGamePlayLogRepository := mockRepository.NewMockGamePlayLogV2(ctrl)
			mockEditionRepository := mockRepository.NewMockEdition(ctrl)
			mockGameRepository := mockRepository.NewMockGameV2(ctrl)
			mockGameVersionRepository := mockRepository.NewMockGameVersionV2(ctrl)

			playLog := NewGamePlayLog(
				mockDB,
				mockGamePlayLogRepository,
				mockEditionRepository,
				mockGameRepository,
				mockGameVersionRepository,
			)

			mockGamePlayLogRepository.
				EXPECT().
				GetGamePlayLog(gomock.Any(), testCase.playLogID).
				Return(testCase.playLog, testCase.GetGamePlayLogErr)
			if testCase.executeDeleteGamePlayLog {
				mockGamePlayLogRepository.
					EXPECT().
					DeleteGamePlayLog(gomock.Any(), testCase.playLogID).
					Return(testCase.deleteGamePlayLogErr)
			}

			err := playLog.DeleteGamePlayLog(
				t.Context(),
				testCase.editionID,
				testCase.gameID,
				testCase.playLogID,
			)

			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestDeleteLongLogs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		deleteLongLogsErr error
		isErr             bool
	}{
		"正常に削除される": {
			deleteLongLogsErr: nil,
			isErr:             false,
		},
		"Repositoryがエラーを返した場合はエラー": {
			deleteLongLogsErr: assert.AnError,
			isErr:             true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockGamePlayLogRepository := mockRepository.NewMockGamePlayLogV2(ctrl)

			gamePlayLogService := NewGamePlayLog(
				nil,
				mockGamePlayLogRepository,
				nil,
				nil,
				nil,
			)

			mockGamePlayLogRepository.
				EXPECT().
				DeleteLongLogs(ctx, 3*time.Hour).
				Return(testCase.deleteLongLogsErr)

			err := gamePlayLogService.DeleteLongLogs(ctx)

			if testCase.isErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
