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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	type test struct {
		description   string
		editionID     values.LauncherVersionID
		gameID        values.GameID
		gameVersionID values.GameVersionID
		startTime     time.Time

		executeGetEdition bool
		getEditionResult  *domain.LauncherVersion
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
	editionID := values.NewLauncherVersionID()
	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()

	questionnaireURL, _ := url.Parse("https://example.com")
	edition := domain.NewLauncherVersionWithQuestionnaire(
		editionID,
		values.NewLauncherVersionName("v1.0.0"),
		values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
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
			editionID:         values.NewLauncherVersionID(),
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
			getEditionErr:     errors.New("error"),
			isErr:             true,
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
			getGameErr:        errors.New("error"),
			isErr:             true,
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
			getGameVersionByIDErr:     errors.New("error"),
			isErr:                     true,
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
			createGamePlayLogErr:      errors.New("error"),
			isErr:                     true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
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
					CreateGamePlayLog(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, playLog *domain.GamePlayLog) error {
						assert.Equal(t, testCase.editionID, playLog.GetEditionID())
						assert.Equal(t, testCase.gameID, playLog.GetGameID())
						assert.Equal(t, testCase.gameVersionID, playLog.GetGameVersionID())
						assert.Equal(t, testCase.startTime, playLog.GetStartTime())
						assert.Nil(t, playLog.GetEndTime())
						assert.NotEqual(t, uuid.Nil, playLog.GetID())
						assert.WithinDuration(t, time.Now(), playLog.GetCreatedAt(), time.Second)
						assert.WithinDuration(t, time.Now(), playLog.GetUpdatedAt(), time.Second)
						return testCase.createGamePlayLogErr
					})
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
	editionID := values.NewLauncherVersionID()
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
			getGamePlayLogErr:     errors.New("error"),
			isErr:                 true,
		},
		{
			description:           "終了時刻が開始時刻より前なのでErrInvalidEndTime",
			playLogID:             playLogID,
			endTime:               now.Add(-2 * time.Hour), // StartTimeより前
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
			updateGamePlayLogEndTimeErr:     errors.New("error"),
			isErr:                           true,
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	sampleStats := &domain.GamePlayStats{
		GameID:           gameID,
		TotalPlayCount:   10,
		TotalPlaySeconds: 3600 * time.Second,
		HourlyStats: []*domain.HourlyPlayStats{
			{
				StartTime: now.Add(-2 * time.Hour).Truncate(time.Hour),
				PlayCount: 5,
				PlayTime:  1800 * time.Second,
			},
			{
				StartTime: now.Add(-1 * time.Hour).Truncate(time.Hour),
				PlayCount: 5,
				PlayTime:  1800 * time.Second,
			},
		},
	}

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
			getGameErr:     errors.New("error"),
			isErr:          true,
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
			getGameVersionByIDErr:     errors.New("error"),
			isErr:                     true,
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
			getGamePlayStatsErr:     errors.New("error"),
			isErr:                   true,
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
			getGamePlayStatsResult: &domain.GamePlayStats{
				GameID:           gameID,
				TotalPlayCount:   0,
				TotalPlaySeconds: 0,
				HourlyStats:      []*domain.HourlyPlayStats{},
			},
			isErr: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
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
				assert.Equal(t, testCase.getGamePlayStatsResult.GameID, stats.GameID)
				assert.Equal(t, testCase.getGamePlayStatsResult.TotalPlayCount, stats.TotalPlayCount)
				assert.Equal(t, testCase.getGamePlayStatsResult.TotalPlaySeconds, stats.TotalPlaySeconds)
				assert.Equal(t, len(testCase.getGamePlayStatsResult.HourlyStats), len(stats.HourlyStats))
			}
		})
	}
}

func TestGetEditionPlayStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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

	type test struct {
		description string
		editionID   values.LauncherVersionID
		start       time.Time
		end         time.Time

		executeGetEdition bool
		getEditionResult  *domain.LauncherVersion
		getEditionErr     error

		executeGetEditionPlayStats bool
		getEditionPlayStatsResult  *domain.EditionPlayStats
		getEditionPlayStatsErr     error

		isErr bool
		err   error
	}

	now := time.Now()
	editionID := values.NewLauncherVersionID()
	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	questionnaireURL, _ := url.Parse("https://example.com")
	edition := domain.NewLauncherVersionWithQuestionnaire(
		editionID,
		values.NewLauncherVersionName("v1.0.0"),
		values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
		now.Add(-24*time.Hour),
	)

	sampleEditionStats := &domain.EditionPlayStats{
		EditionID:        editionID,
		EditionName:      values.NewLauncherVersionName(""), // これは後でサービスが設定する
		TotalPlayCount:   15,
		TotalPlaySeconds: 5400 * time.Second,
		GameStats: []*domain.GamePlayStatsInEdition{
			{
				GameID:    gameID1,
				PlayCount: 8,
				PlayTime:  3200 * time.Second,
			},
			{
				GameID:    gameID2,
				PlayCount: 7,
				PlayTime:  2200 * time.Second,
			},
		},
		HourlyStats: []*domain.HourlyPlayStats{
			{
				StartTime: now.Add(-3 * time.Hour).Truncate(time.Hour),
				PlayCount: 5,
				PlayTime:  1800 * time.Second,
			},
			{
				StartTime: now.Add(-2 * time.Hour).Truncate(time.Hour),
				PlayCount: 5,
				PlayTime:  1800 * time.Second,
			},
			{
				StartTime: now.Add(-1 * time.Hour).Truncate(time.Hour),
				PlayCount: 5,
				PlayTime:  1800 * time.Second,
			},
		},
	}

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
			editionID:         values.NewLauncherVersionID(),
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
			getEditionErr:     errors.New("error"),
			isErr:             true,
		},
		{
			description:                "GetEditionPlayStatsがエラーなのでエラー",
			editionID:                  editionID,
			start:                      now.Add(-24 * time.Hour),
			end:                        now,
			executeGetEdition:          true,
			getEditionResult:           edition,
			executeGetEditionPlayStats: true,
			getEditionPlayStatsErr:     errors.New("error"),
			isErr:                      true,
		},
		{
			description:                "開始時刻と終了時刻が同じでも正常に取得できる",
			editionID:                  editionID,
			start:                      now,
			end:                        now,
			executeGetEdition:          true,
			getEditionResult:           edition,
			executeGetEditionPlayStats: true,
			getEditionPlayStatsResult: &domain.EditionPlayStats{
				EditionID:        editionID,
				EditionName:      values.NewLauncherVersionName(""),
				TotalPlayCount:   0,
				TotalPlaySeconds: 0,
				GameStats:        []*domain.GamePlayStatsInEdition{},
				HourlyStats:      []*domain.HourlyPlayStats{},
			},
			isErr: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
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
				assert.Equal(t, testCase.editionID, stats.EditionID)
				assert.Equal(t, testCase.getEditionPlayStatsResult.TotalPlayCount, stats.TotalPlayCount)
				assert.Equal(t, testCase.getEditionPlayStatsResult.TotalPlaySeconds, stats.TotalPlaySeconds)
				assert.Equal(t, len(testCase.getEditionPlayStatsResult.GameStats), len(stats.GameStats))
				assert.Equal(t, len(testCase.getEditionPlayStatsResult.HourlyStats), len(stats.HourlyStats))
				// サービスがエディション名を設定することを確認
				assert.Equal(t, testCase.getEditionResult.GetName(), stats.EditionName)
			}
		})
	}
}
