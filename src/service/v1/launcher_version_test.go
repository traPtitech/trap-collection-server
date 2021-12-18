package v1

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/mock"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func TestCreateLauncherVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockGameRepository := mock.NewMockGame(ctrl)

	launcherVersionService := NewLauncherVersion(mockDB, mockLauncherVersionRepository, mockGameRepository)

	type test struct {
		description              string
		name                     values.LauncherVersionName
		questionnaireURL         values.LauncherVersionQuestionnaireURL
		CreateLauncherVersionErr error
		isErr                    bool
		err                      error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	testCases := []test{
		{
			description:      "questionnaireURLありでエラーなし",
			name:             values.NewLauncherVersionName("name"),
			questionnaireURL: values.NewLauncherVersionQuestionnaireURL(urlLink),
		},
		{
			description: "questionnaireURLなしでエラーなし",
			name:        values.NewLauncherVersionName("name"),
		},
		{
			description:              "CreateLauncherVersionがエラーなのでエラー",
			name:                     values.NewLauncherVersionName("name"),
			questionnaireURL:         values.NewLauncherVersionQuestionnaireURL(urlLink),
			CreateLauncherVersionErr: errors.New("error"),
			isErr:                    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				CreateLauncherVersion(ctx, gomock.Any()).
				Return(testCase.CreateLauncherVersionErr)

			launcherVersion, err := launcherVersionService.CreateLauncherVersion(ctx, testCase.name, testCase.questionnaireURL)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.name, launcherVersion.GetName())
			assert.WithinDuration(t, time.Now(), launcherVersion.GetCreatedAt(), 2*time.Second)

			questionnaireURL, err := launcherVersion.GetQuestionnaireURL()

			if testCase.questionnaireURL == nil {
				assert.True(t, errors.Is(err, domain.ErrNoQuestionnaire))
			} else {
				assert.Equal(t, testCase.questionnaireURL, questionnaireURL)
			}
		})
	}
}

func TestGetLauncherVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockGameRepository := mock.NewMockGame(ctrl)

	launcherVersionService := NewLauncherVersion(mockDB, mockLauncherVersionRepository, mockGameRepository)

	type test struct {
		description            string
		launcherVersions       []*domain.LauncherVersion
		GetLauncherVersionsErr error
		isErr                  bool
		err                    error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					values.NewLauncherVersionID(),
					values.NewLauncherVersionName("name"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					time.Now(),
				),
			},
		},
		{
			description:            "GetLauncherVersionsがエラーなのでエラー",
			GetLauncherVersionsErr: errors.New("error"),
			isErr:                  true,
		},
		{
			description:      "launcherVersionsが空でもエラーなし",
			launcherVersions: []*domain.LauncherVersion{},
		},
		{
			description: "launcherVersionsが複数でもエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					values.NewLauncherVersionID(),
					values.NewLauncherVersionName("name"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					time.Now(),
				),
				domain.NewLauncherVersionWithQuestionnaire(
					values.NewLauncherVersionID(),
					values.NewLauncherVersionName("name"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					time.Now(),
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				GetLauncherVersions(ctx).
				Return(testCase.launcherVersions, testCase.GetLauncherVersionsErr)

			launcherVersions, err := launcherVersionService.GetLauncherVersions(ctx)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Len(t, launcherVersions, len(testCase.launcherVersions))

			for i, launcherVersion := range launcherVersions {
				assert.Equal(t, testCase.launcherVersions[i].GetName(), launcherVersion.GetName())
				assert.WithinDuration(t, testCase.launcherVersions[i].GetCreatedAt(), launcherVersion.GetCreatedAt(), 2*time.Second)

				questionnaireURL, err := launcherVersion.GetQuestionnaireURL()

				if errors.Is(err, domain.ErrNoQuestionnaire) {
					_, err = testCase.launcherVersions[i].GetQuestionnaireURL()
					assert.True(t, errors.Is(err, domain.ErrNoQuestionnaire))
				} else {
					expectQuestionnaireURL, err := testCase.launcherVersions[i].GetQuestionnaireURL()
					assert.False(t, errors.Is(err, domain.ErrNoQuestionnaire))
					assert.Equal(t, expectQuestionnaireURL, questionnaireURL)
				}
			}
		})
	}
}

func TestGetLauncherVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockGameRepository := mock.NewMockGame(ctrl)

	launcherVersionService := NewLauncherVersion(mockDB, mockLauncherVersionRepository, mockGameRepository)

	type test struct {
		description                     string
		launcherVersionID               values.LauncherVersionID
		launcherVersion                 *domain.LauncherVersion
		GetLauncherVersionErr           error
		executeGetGameByLauncherVersion bool
		games                           []*domain.Game
		GetGamesByLauncherVersionErr    error
		isErr                           bool
		err                             error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	launcherVersionID := values.NewLauncherVersionID()

	testCases := []test{
		{
			description:       "特に問題ないのでエラーなし",
			launcherVersionID: launcherVersionID,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameByLauncherVersion: true,
			games: []*domain.Game{
				domain.NewGame(
					values.NewGameID(),
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
			},
		},
		{
			description:           "GetLauncherVersionがErrRecordNotFoundなのでErrNoLauncherVersion",
			launcherVersionID:     launcherVersionID,
			GetLauncherVersionErr: repository.ErrRecordNotFound,
			isErr:                 true,
			err:                   service.ErrNoLauncherVersion,
		},
		{
			description:           "GetLauncherVersionがエラーなのでエラー",
			launcherVersionID:     launcherVersionID,
			GetLauncherVersionErr: errors.New("error"),
			isErr:                 true,
		},
		{
			description:       "GetGamesByLauncherVersionがエラーなのでエラー",
			launcherVersionID: launcherVersionID,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameByLauncherVersion: true,
			GetGamesByLauncherVersionErr:    errors.New("error"),
			isErr:                           true,
		},
		{
			description:       "gameが0個でもエラーなし",
			launcherVersionID: launcherVersionID,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameByLauncherVersion: true,
			games:                           []*domain.Game{},
		},
		{
			description:       "gameが複数でもエラーなし",
			launcherVersionID: launcherVersionID,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameByLauncherVersion: true,
			games: []*domain.Game{
				domain.NewGame(
					values.NewGameID(),
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
				domain.NewGame(
					values.NewGameID(),
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				GetLauncherVersion(ctx, testCase.launcherVersionID, repository.LockTypeNone).
				Return(testCase.launcherVersion, testCase.GetLauncherVersionErr)

			if testCase.executeGetGameByLauncherVersion {
				mockGameRepository.
					EXPECT().
					GetGamesByLauncherVersion(ctx, testCase.launcherVersionID).
					Return(testCase.games, testCase.GetGamesByLauncherVersionErr)
			}

			launcherVersion, games, err := launcherVersionService.GetLauncherVersion(ctx, testCase.launcherVersionID)

			if testCase.isErr {
				if testCase.err == nil {
					assert.Error(t, err)
				} else if !errors.Is(err, testCase.err) {
					t.Errorf("error must be %v, but actual is %v", testCase.err, err)
				}
			} else {
				assert.NoError(t, err)
			}
			if err != nil {
				return
			}

			assert.Equal(t, testCase.launcherVersion.GetID(), launcherVersion.GetID())
			assert.Equal(t, testCase.launcherVersion.GetName(), launcherVersion.GetName())
			assert.WithinDuration(t, testCase.launcherVersion.GetCreatedAt(), launcherVersion.GetCreatedAt(), time.Second)

			questionnaireURL, err := launcherVersion.GetQuestionnaireURL()

			if errors.Is(err, domain.ErrNoQuestionnaire) {
				_, err = testCase.launcherVersion.GetQuestionnaireURL()
				assert.True(t, errors.Is(err, domain.ErrNoQuestionnaire))
			} else {
				expectQuestionnaireURL, err := testCase.launcherVersion.GetQuestionnaireURL()
				assert.False(t, errors.Is(err, domain.ErrNoQuestionnaire))
				assert.Equal(t, expectQuestionnaireURL, questionnaireURL)
			}

			for i, game := range games {
				assert.Equal(t, testCase.games[i].GetID(), game.GetID())
				assert.Equal(t, testCase.games[i].GetName(), game.GetName())
				assert.Equal(t, testCase.games[i].GetDescription(), game.GetDescription())
				assert.WithinDuration(t, testCase.games[i].GetCreatedAt(), game.GetCreatedAt(), time.Second)
			}
		})
	}
}

func TestAddGamesToLauncherVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockGameRepository := mock.NewMockGame(ctrl)

	launcherVersionService := NewLauncherVersion(mockDB, mockLauncherVersionRepository, mockGameRepository)

	type test struct {
		description                      string
		launcherVersionID                values.LauncherVersionID
		gameIDs                          []values.GameID
		launcherVersion                  *domain.LauncherVersion
		GetLauncherVersionErr            error
		executeGetGamesByIDs             bool
		games                            []*domain.Game
		GetGamesByIDsErr                 error
		executeAddGamesToLauncherVersion bool
		AddGamesToLauncherVersionErr     error
		isErr                            bool
		err                              error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	launcherVersionID := values.NewLauncherVersionID()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	testCases := []test{
		{
			description:       "特に問題ないのでエラーなし",
			launcherVersionID: launcherVersionID,
			gameIDs: []values.GameID{
				gameID1,
			},
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGamesByIDs: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
			},
			executeAddGamesToLauncherVersion: true,
		},
		{
			description:       "gameが複数でもエラーなし",
			launcherVersionID: launcherVersionID,
			gameIDs: []values.GameID{
				gameID1,
				gameID2,
			},
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGamesByIDs: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
				domain.NewGame(
					gameID2,
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
			},
			executeAddGamesToLauncherVersion: true,
		},
		{
			description:       "gameIDがからでもエラーなし",
			launcherVersionID: launcherVersionID,
			gameIDs:           []values.GameID{},
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGamesByIDs:             true,
			games:                            []*domain.Game{},
			executeAddGamesToLauncherVersion: true,
		},
		{
			description:       "LauncherVersionが存在しないのでErrNoLauncherVersion",
			launcherVersionID: launcherVersionID,
			gameIDs: []values.GameID{
				gameID1,
			},
			GetLauncherVersionErr: repository.ErrRecordNotFound,
			isErr:                 true,
			err:                   service.ErrNoLauncherVersion,
		},
		{
			description:       "GetLauncherVersionがエラーなのでエラー",
			launcherVersionID: launcherVersionID,
			gameIDs: []values.GameID{
				gameID1,
			},
			GetLauncherVersionErr: errors.New("error"),
			isErr:                 true,
		},
		{
			description:       "GetGamesByIDsがエラーなのでエラー",
			launcherVersionID: launcherVersionID,
			gameIDs: []values.GameID{
				gameID1,
			},
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGamesByIDs: true,
			GetGamesByIDsErr:     errors.New("error"),
			isErr:                true,
		},
		{
			description:       "存在しないGameIDが含まれるのでErrNoGame",
			launcherVersionID: launcherVersionID,
			gameIDs: []values.GameID{
				gameID1,
				gameID2,
			},
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGamesByIDs: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
			},
			isErr: true,
			err:   service.ErrNoGame,
		},
		{
			description:       "AddGamesToLauncherVersionがエラーなのでエラー",
			launcherVersionID: launcherVersionID,
			gameIDs: []values.GameID{
				gameID1,
			},
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGamesByIDs: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("name"),
					values.NewGameDescription("description"),
					time.Now(),
				),
			},
			executeAddGamesToLauncherVersion: true,
			AddGamesToLauncherVersionErr:     errors.New("error"),
			isErr:                            true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				GetLauncherVersion(gomock.Any(), testCase.launcherVersionID, repository.LockTypeRecord).
				Return(testCase.launcherVersion, testCase.GetLauncherVersionErr)

			if testCase.executeGetGamesByIDs {
				mockGameRepository.
					EXPECT().
					GetGamesByIDs(gomock.Any(), testCase.gameIDs, repository.LockTypeRecord).
					Return(testCase.games, testCase.GetGamesByIDsErr)
			}

			if testCase.executeAddGamesToLauncherVersion {
				mockLauncherVersionRepository.
					EXPECT().
					AddGamesToLauncherVersion(gomock.Any(), testCase.launcherVersionID, testCase.gameIDs).
					Return(testCase.AddGamesToLauncherVersionErr)
			}

			err := launcherVersionService.AddGamesToLauncherVersion(ctx, testCase.launcherVersionID, testCase.gameIDs)

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
