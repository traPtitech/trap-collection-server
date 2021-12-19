package v1

import (
	"context"
	"errors"
	"net/url"
	"strings"
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

func TestGetLauncherVersionCheckList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockDB(ctrl)
	mockLauncherVersionRepository := mock.NewMockLauncherVersion(ctrl)
	mockGameRepository := mock.NewMockGame(ctrl)

	launcherVersionService := NewLauncherVersion(mockDB, mockLauncherVersionRepository, mockGameRepository)

	type test struct {
		description                          string
		launcherVersionID                    values.LauncherVersionID
		env                                  *values.LauncherEnvironment
		launcherVersion                      *domain.LauncherVersion
		GetLauncherVersionErr                error
		executeGetGameInfosByLauncherVersion bool
		gameInfos                            []*repository.GameInfo
		GetGameInfosByLauncherVersionErr     error
		checkList                            []*service.CheckListItem
		isErr                                bool
		err                                  error
	}

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse url: %v", err)
	}

	launcherVersionID := values.NewLauncherVersionID()
	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()
	gameURLID1 := values.NewGameURLID()
	gameFileID1 := values.NewGameFileID()
	gameFileID2 := values.NewGameFileID()
	gameImageID1 := values.NewGameImageID()
	gameImageID2 := values.NewGameImageID()
	gameVideoID1 := values.NewGameVideoID()
	gameVideoID2 := values.NewGameVideoID()

	now := time.Now()

	hash, err := values.NewGameFileHash(strings.NewReader("hash"))
	if err != nil {
		t.Fatalf("failed to create game file hash: %v", err)
	}

	testCases := []test{
		{
			description:       "特に問題ないのでエラーなし",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID1,
						values.GameFileTypeJar,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
		{
			description:       "urlのゲームがあってもエラーなし",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestURL: domain.NewGameURL(
						gameURLID1,
						values.NewGameURLLink(urlLink),
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestURL: domain.NewGameURL(
						gameURLID1,
						values.NewGameURLLink(urlLink),
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
		{
			description:       "fileが存在する場合はurlは無視",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestURL: domain.NewGameURL(
						gameURLID1,
						values.NewGameURLLink(urlLink),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID1,
						values.GameFileTypeJar,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
		{
			description:           "GetLauncherVersionがRecordNotFoundなのでErrNoLauncherVersion",
			launcherVersionID:     launcherVersionID,
			env:                   values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			GetLauncherVersionErr: repository.ErrRecordNotFound,
			isErr:                 true,
			err:                   service.ErrNoLauncherVersion,
		},
		{
			description:           "GetLauncherVersionがエラーなのでエラー",
			launcherVersionID:     launcherVersionID,
			env:                   values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			GetLauncherVersionErr: errors.New("error"),
			isErr:                 true,
		},
		{
			description:       "GetGameInfosByLauncherVersionがエラーなのでエラー",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			GetGameInfosByLauncherVersionErr:     errors.New("error"),
			isErr:                                true,
		},
		{
			description:       "Fileが存在しなくてもエラーなし",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{},
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: nil,
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
		{
			description:       "imageが存在しなくてもエラーなし",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: nil,
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID1,
						values.GameFileTypeJar,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: nil,
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
		{
			description:       "videoが存在しなくてもエラーなし",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: nil,
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID1,
						values.GameFileTypeJar,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: nil,
				},
			},
		},
		{
			description:       "gameが存在しなくてもエラーなし",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos:                            []*repository.GameInfo{},
			checkList:                            []*service.CheckListItem{},
		},
		{
			description:       "ゲームが複数でもエラーなし",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
				{
					Game: domain.NewGame(
						gameID2,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID2,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID2,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID2,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID2,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID1,
						values.GameFileTypeJar,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
				{
					Game: domain.NewGame(
						gameID2,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID2,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID2,
						values.GameFileTypeJar,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID2,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID2,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
		{
			description:       "ファイルが複数でwindows用のものがある時、そちらを優先",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSWindows),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
						domain.NewGameFile(
							gameFileID2,
							values.GameFileTypeWindows,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID2,
						values.GameFileTypeWindows,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
		{
			description:       "ファイルが複数でmac用のものがある時、そちらを優先",
			launcherVersionID: launcherVersionID,
			env:               values.NewLauncherEnvironment(values.LauncherEnvironmentOSMac),
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("name"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				time.Now(),
			),
			executeGetGameInfosByLauncherVersion: true,
			gameInfos: []*repository.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFiles: []*domain.GameFile{
						domain.NewGameFile(
							gameFileID1,
							values.GameFileTypeJar,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
						domain.NewGameFile(
							gameFileID2,
							values.GameFileTypeMac,
							values.NewGameFileEntryPoint("entryPoint"),
							hash,
							now,
						),
					},
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
			checkList: []*service.CheckListItem{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("name"),
						values.NewGameDescription("description"),
						time.Now(),
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("name"),
						values.NewGameVersionDescription("description"),
						time.Now(),
					),
					LatestFile: domain.NewGameFile(
						gameFileID2,
						values.GameFileTypeMac,
						values.NewGameFileEntryPoint("entryPoint"),
						hash,
						now,
					),
					LatestImage: domain.NewGameImage(
						gameImageID1,
						values.GameImageTypeJpeg,
						time.Now(),
					),
					LatestVideo: domain.NewGameVideo(
						gameVideoID1,
						values.GameVideoTypeMp4,
						time.Now(),
					),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionRepository.
				EXPECT().
				GetLauncherVersion(gomock.Any(), launcherVersionID, repository.LockTypeNone).
				Return(testCase.launcherVersion, testCase.GetLauncherVersionErr)

			if testCase.executeGetGameInfosByLauncherVersion {
				mockGameRepository.
					EXPECT().
					GetGameInfosByLauncherVersion(gomock.Any(), launcherVersionID, gomock.Any()).
					Return(testCase.gameInfos, testCase.GetGameInfosByLauncherVersionErr)
			}

			checkList, err := launcherVersionService.GetLauncherVersionCheckList(
				ctx,
				testCase.launcherVersionID,
				testCase.env,
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
			if err != nil {
				return
			}

			assert.Len(t, checkList, len(testCase.checkList))

			for i, checkListItem := range checkList {
				assert.Equal(t, testCase.checkList[i].Game.GetID(), checkListItem.Game.GetID())
				assert.Equal(t, testCase.checkList[i].Game.GetName(), checkListItem.Game.GetName())
				assert.Equal(t, testCase.checkList[i].Game.GetDescription(), checkListItem.Game.GetDescription())
				assert.WithinDuration(t, testCase.checkList[i].Game.GetCreatedAt(), checkListItem.Game.GetCreatedAt(), time.Second)

				assert.Equal(t, testCase.checkList[i].LatestVersion.GetID(), checkListItem.LatestVersion.GetID())
				assert.Equal(t, testCase.checkList[i].LatestVersion.GetName(), checkListItem.LatestVersion.GetName())
				assert.Equal(t, testCase.checkList[i].LatestVersion.GetDescription(), checkListItem.LatestVersion.GetDescription())
				assert.WithinDuration(t, testCase.checkList[i].LatestVersion.GetCreatedAt(), checkListItem.LatestVersion.GetCreatedAt(), time.Second)

				if testCase.checkList[i].LatestFile == nil {
					assert.Nil(t, checkListItem.LatestFile)
				} else {
					assert.Equal(t, testCase.checkList[i].LatestFile.GetID(), checkListItem.LatestFile.GetID())
					assert.Equal(t, testCase.checkList[i].LatestFile.GetFileType(), checkListItem.LatestFile.GetFileType())
					assert.Equal(t, testCase.checkList[i].LatestFile.GetEntryPoint(), checkListItem.LatestFile.GetEntryPoint())
					assert.Equal(t, testCase.checkList[i].LatestFile.GetHash(), checkListItem.LatestFile.GetHash())
				}

				if testCase.checkList[i].LatestImage == nil {
					assert.Nil(t, checkListItem.LatestImage)
				} else {
					assert.Equal(t, testCase.checkList[i].LatestImage.GetID(), checkListItem.LatestImage.GetID())
					assert.Equal(t, testCase.checkList[i].LatestImage.GetType(), checkListItem.LatestImage.GetType())
					assert.WithinDuration(t, testCase.checkList[i].LatestImage.GetCreatedAt(), checkListItem.LatestImage.GetCreatedAt(), time.Second)
				}

				if testCase.checkList[i].LatestVideo == nil {
					assert.Nil(t, checkListItem.LatestVideo)
				} else {
					assert.Equal(t, testCase.checkList[i].LatestVideo.GetID(), checkListItem.LatestVideo.GetID())
					assert.Equal(t, testCase.checkList[i].LatestVideo.GetType(), checkListItem.LatestVideo.GetType())
					assert.WithinDuration(t, testCase.checkList[i].LatestVideo.GetCreatedAt(), checkListItem.LatestVideo.GetCreatedAt(), time.Second)
				}
			}
		})
	}
}
