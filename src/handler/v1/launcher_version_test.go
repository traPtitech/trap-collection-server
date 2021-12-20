package v1

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetVersions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherVersionService := mock.NewMockLauncherVersion(ctrl)

	launcherVersionHandler := NewLauncherVersion(mockLauncherVersionService)

	type test struct {
		description            string
		launcherVersions       []*domain.LauncherVersion
		GetLauncherVersionsErr error
		expect                 []*openapi.Version
		isErr                  bool
		err                    error
		statusCode             int
	}

	launcherVersionID1 := values.NewLauncherVersionID()
	launcherVersionID2 := values.NewLauncherVersionID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithoutQuestionnaire(
					launcherVersionID1,
					values.NewLauncherVersionName("2020.1.1"),
					now,
				),
			},
			expect: []*openapi.Version{
				{
					Id:        uuid.UUID(launcherVersionID1).String(),
					Name:      "2020.1.1",
					AnkeTo:    "",
					CreatedAt: now,
				},
			},
		},
		{
			description:            "GetLauncherVersionsがエラーなので500",
			GetLauncherVersionsErr: errors.New("GetLauncherVersions error"),
			isErr:                  true,
			statusCode:             http.StatusInternalServerError,
		},
		{
			description: "Questionnaireありのランチャーバージョンでもエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithQuestionnaire(
					launcherVersionID1,
					values.NewLauncherVersionName("2020.1.1"),
					values.NewLauncherVersionQuestionnaireURL(urlLink),
					now,
				),
			},
			expect: []*openapi.Version{
				{
					Id:        uuid.UUID(launcherVersionID1).String(),
					Name:      "2020.1.1",
					AnkeTo:    "https://example.com",
					CreatedAt: now,
				},
			},
		},
		{
			description:      "ランチャーバージョンがなくてもエラーなし",
			launcherVersions: []*domain.LauncherVersion{},
			expect:           []*openapi.Version{},
		},
		{
			description: "ランチャーバージョンが複数でもエラーなし",
			launcherVersions: []*domain.LauncherVersion{
				domain.NewLauncherVersionWithoutQuestionnaire(
					launcherVersionID1,
					values.NewLauncherVersionName("2020.1.1"),
					now,
				),
				domain.NewLauncherVersionWithoutQuestionnaire(
					launcherVersionID2,
					values.NewLauncherVersionName("2020.1.1"),
					now,
				),
			},
			expect: []*openapi.Version{
				{
					Id:        uuid.UUID(launcherVersionID1).String(),
					Name:      "2020.1.1",
					AnkeTo:    "",
					CreatedAt: now,
				},
				{
					Id:        uuid.UUID(launcherVersionID2).String(),
					Name:      "2020.1.1",
					AnkeTo:    "",
					CreatedAt: now,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			mockLauncherVersionService.
				EXPECT().
				GetLauncherVersions(gomock.Any()).
				Return(testCase.launcherVersions, testCase.GetLauncherVersionsErr)

			launcherVersions, err := launcherVersionHandler.GetVersions()

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Len(t, launcherVersions, len(testCase.expect))

			for i, expect := range testCase.expect {
				assert.Equal(t, *expect, *launcherVersions[i])
			}
		})
	}
}

func TestPostVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherVersionService := mock.NewMockLauncherVersion(ctrl)

	launcherVersionHandler := NewLauncherVersion(mockLauncherVersionService)

	type test struct {
		description                  string
		newVersion                   *openapi.NewVersion
		version                      *domain.LauncherVersion
		executeCreateLauncherVersion bool
		CreateLauncherVersionErr     error
		expect                       *openapi.VersionMeta
		isErr                        bool
		err                          error
		statusCode                   int
	}

	launcherVersionID := values.NewLauncherVersionID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description: "エラーなしなので問題なし",
			newVersion: &openapi.NewVersion{
				Name:   "2020.1.1",
				AnkeTo: "https://example.com",
			},
			executeCreateLauncherVersion: true,
			version: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("2020.1.1"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				now,
			),
			expect: &openapi.VersionMeta{
				Id:        uuid.UUID(launcherVersionID).String(),
				Name:      "2020.1.1",
				AnkeTo:    "https://example.com",
				CreatedAt: now,
			},
		},
		{
			description: "nameが空なので400",
			newVersion: &openapi.NewVersion{
				Name:   "",
				AnkeTo: "https://example.com",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "nameが長すぎるので400",
			newVersion: &openapi.NewVersion{
				Name:   "2020.1.1-012345678901234567890123",
				AnkeTo: "https://example.com",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "アンケートなしでも問題なし",
			newVersion: &openapi.NewVersion{
				Name:   "2020.1.1",
				AnkeTo: "",
			},
			executeCreateLauncherVersion: true,
			version: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID,
				values.NewLauncherVersionName("2020.1.1"),
				now,
			),
			expect: &openapi.VersionMeta{
				Id:        uuid.UUID(launcherVersionID).String(),
				Name:      "2020.1.1",
				AnkeTo:    "",
				CreatedAt: now,
			},
		},
		{
			description: "アンケートのURLが不正なので400",
			newVersion: &openapi.NewVersion{
				Name:   "2020.1.1",
				AnkeTo: " https://example.com",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "CreateLauncherVersionがエラーなので500",
			newVersion: &openapi.NewVersion{
				Name:   "2020.1.1",
				AnkeTo: "https://example.com",
			},
			executeCreateLauncherVersion: true,
			CreateLauncherVersionErr:     errors.New("error"),
			isErr:                        true,
			statusCode:                   http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeCreateLauncherVersion {
				mockLauncherVersionService.
					EXPECT().
					CreateLauncherVersion(gomock.Any(), values.NewLauncherVersionName(testCase.newVersion.Name), gomock.Any()).
					Return(testCase.version, testCase.CreateLauncherVersionErr)
			}

			launcherVersion, err := launcherVersionHandler.PostVersion(testCase.newVersion)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, *testCase.expect, *launcherVersion)
		})
	}
}

func TestGetVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherVersionService := mock.NewMockLauncherVersion(ctrl)

	launcherVersionHandler := NewLauncherVersion(mockLauncherVersionService)

	type test struct {
		description               string
		strLauncherVersionID      string
		executeGetLauncherVersion bool
		launcherVersionID         values.LauncherVersionID
		launcherVersion           *domain.LauncherVersion
		games                     []*domain.Game
		GetLauncherVersionErr     error
		expect                    *openapi.VersionDetails
		isErr                     bool
		err                       error
		statusCode                int
	}

	launcherVersionID1 := values.NewLauncherVersionID()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description:               "エラーなしなので問題なし",
			strLauncherVersionID:      uuid.UUID(launcherVersionID1).String(),
			executeGetLauncherVersion: true,
			launcherVersionID:         launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("2020.1.1"),
				now,
			),
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("game1"),
					values.NewGameDescription("description1"),
					now,
				),
			},
			expect: &openapi.VersionDetails{
				Id:        uuid.UUID(launcherVersionID1).String(),
				Name:      "2020.1.1",
				AnkeTo:    "",
				CreatedAt: now,
				Games: []openapi.GameMeta{
					{
						Id:   uuid.UUID(gameID1).String(),
						Name: "game1",
					},
				},
			},
		},
		{
			description:          "ランチャーバージョンIDが不正なので400",
			strLauncherVersionID: "invalid",
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description:               "ランチャーバージョンが存在しないので400",
			strLauncherVersionID:      uuid.UUID(launcherVersionID1).String(),
			executeGetLauncherVersion: true,
			launcherVersionID:         launcherVersionID1,
			GetLauncherVersionErr:     service.ErrNoLauncherVersion,
			isErr:                     true,
			statusCode:                http.StatusBadRequest,
		},
		{
			description:               "GetLauncherVersionがエラーなので500",
			strLauncherVersionID:      uuid.UUID(launcherVersionID1).String(),
			executeGetLauncherVersion: true,
			launcherVersionID:         launcherVersionID1,
			GetLauncherVersionErr:     errors.New("error"),
			isErr:                     true,
			statusCode:                http.StatusInternalServerError,
		},
		{
			description:               "アンケートありでも問題なし",
			strLauncherVersionID:      uuid.UUID(launcherVersionID1).String(),
			executeGetLauncherVersion: true,
			launcherVersionID:         launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("2020.1.1"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				now,
			),
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("game1"),
					values.NewGameDescription("description1"),
					now,
				),
			},
			expect: &openapi.VersionDetails{
				Id:        uuid.UUID(launcherVersionID1).String(),
				Name:      "2020.1.1",
				AnkeTo:    "https://example.com",
				CreatedAt: now,
				Games: []openapi.GameMeta{
					{
						Id:   uuid.UUID(gameID1).String(),
						Name: "game1",
					},
				},
			},
		},
		{
			description:               "ゲームがなくても問題なし",
			strLauncherVersionID:      uuid.UUID(launcherVersionID1).String(),
			executeGetLauncherVersion: true,
			launcherVersionID:         launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("2020.1.1"),
				now,
			),
			games: []*domain.Game{},
			expect: &openapi.VersionDetails{
				Id:        uuid.UUID(launcherVersionID1).String(),
				Name:      "2020.1.1",
				AnkeTo:    "",
				CreatedAt: now,
				Games:     []openapi.GameMeta{},
			},
		},
		{
			description:               "ゲームが複数でも問題なし",
			strLauncherVersionID:      uuid.UUID(launcherVersionID1).String(),
			executeGetLauncherVersion: true,
			launcherVersionID:         launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("2020.1.1"),
				now,
			),
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("game1"),
					values.NewGameDescription("description1"),
					now,
				),
				domain.NewGame(
					gameID2,
					values.NewGameName("game2"),
					values.NewGameDescription("description2"),
					now,
				),
			},
			expect: &openapi.VersionDetails{
				Id:        uuid.UUID(launcherVersionID1).String(),
				Name:      "2020.1.1",
				AnkeTo:    "",
				CreatedAt: now,
				Games: []openapi.GameMeta{
					{
						Id:   uuid.UUID(gameID1).String(),
						Name: "game1",
					},
					{
						Id:   uuid.UUID(gameID2).String(),
						Name: "game2",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeGetLauncherVersion {
				mockLauncherVersionService.
					EXPECT().
					GetLauncherVersion(gomock.Any(), testCase.launcherVersionID).
					Return(testCase.launcherVersion, testCase.games, testCase.GetLauncherVersionErr)
			}

			launcherVersion, err := launcherVersionHandler.GetVersion(testCase.strLauncherVersionID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, testCase.expect.Id, launcherVersion.Id)
			assert.Equal(t, testCase.expect.Name, launcherVersion.Name)
			assert.Equal(t, testCase.expect.AnkeTo, launcherVersion.AnkeTo)
			assert.WithinDuration(t, testCase.expect.CreatedAt, launcherVersion.CreatedAt, time.Second)

			assert.Len(t, launcherVersion.Games, len(testCase.expect.Games))

			for i, expect := range testCase.expect.Games {
				assert.Equal(t, expect, launcherVersion.Games[i])
			}
		})
	}
}

func TestPostGameToVersion(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLauncherVersionService := mock.NewMockLauncherVersion(ctrl)

	launcherVersionHandler := NewLauncherVersion(mockLauncherVersionService)

	type test struct {
		description                      string
		strLauncherVersionID             string
		apiGameIDs                       *openapi.GameIDs
		executeAddGamesToLauncherVersion bool
		launcherVersionID                values.LauncherVersionID
		launcherVersion                  *domain.LauncherVersion
		games                            []*domain.Game
		AddGamesToLauncherVersionErr     error
		expect                           *openapi.VersionDetails
		isErr                            bool
		err                              error
		statusCode                       int
	}

	launcherVersionID1 := values.NewLauncherVersionID()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	now := time.Now()

	testCases := []test{
		{
			description:          "エラーなしなので問題なし",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String()},
			},
			executeAddGamesToLauncherVersion: true,
			launcherVersionID:                launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("2020.1.1"),
				now,
			),
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("game1"),
					values.NewGameDescription("description1"),
					now,
				),
			},
			expect: &openapi.VersionDetails{
				Id:        uuid.UUID(launcherVersionID1).String(),
				Name:      "2020.1.1",
				AnkeTo:    "",
				CreatedAt: now,
				Games: []openapi.GameMeta{
					{
						Id:   uuid.UUID(gameID1).String(),
						Name: "game1",
					},
				},
			},
		},
		{
			description:          "ランチャーバージョンIDが不正なので400",
			strLauncherVersionID: "invalid",
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String()},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:          "ゲームIDが不正なので400",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{"invalid"},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:          "ランチャーバージョンが存在しないので400",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String()},
			},
			executeAddGamesToLauncherVersion: true,
			launcherVersionID:                launcherVersionID1,
			AddGamesToLauncherVersionErr:     service.ErrNoLauncherVersion,
			isErr:                            true,
			statusCode:                       http.StatusBadRequest,
		},
		{
			description:          "ゲームが存在しないので400",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String()},
			},
			executeAddGamesToLauncherVersion: true,
			launcherVersionID:                launcherVersionID1,
			AddGamesToLauncherVersionErr:     service.ErrNoGame,
			isErr:                            true,
			statusCode:                       http.StatusBadRequest,
		},
		{
			description:          "ゲームが既に登録されているので400",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String()},
			},
			executeAddGamesToLauncherVersion: true,
			launcherVersionID:                launcherVersionID1,
			AddGamesToLauncherVersionErr:     service.ErrDuplicateGame,
			isErr:                            true,
			statusCode:                       http.StatusBadRequest,
		},
		{
			description:          "GetLauncherVersionがエラーなので500",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String()},
			},
			executeAddGamesToLauncherVersion: true,
			launcherVersionID:                launcherVersionID1,
			AddGamesToLauncherVersionErr:     errors.New("error"),
			isErr:                            true,
			statusCode:                       http.StatusInternalServerError,
		},
		{
			description:          "アンケートありでも問題なし",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String()},
			},
			executeAddGamesToLauncherVersion: true,
			launcherVersionID:                launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("2020.1.1"),
				values.NewLauncherVersionQuestionnaireURL(urlLink),
				now,
			),
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("game1"),
					values.NewGameDescription("description1"),
					now,
				),
			},
			expect: &openapi.VersionDetails{
				Id:        uuid.UUID(launcherVersionID1).String(),
				Name:      "2020.1.1",
				AnkeTo:    "https://example.com",
				CreatedAt: now,
				Games: []openapi.GameMeta{
					{
						Id:   uuid.UUID(gameID1).String(),
						Name: "game1",
					},
				},
			},
		},
		{
			description:          "ゲームが複数でも問題なし",
			strLauncherVersionID: uuid.UUID(launcherVersionID1).String(),
			apiGameIDs: &openapi.GameIDs{
				GameIDs: []string{uuid.UUID(gameID1).String(), uuid.UUID(gameID2).String()},
			},
			executeAddGamesToLauncherVersion: true,
			launcherVersionID:                launcherVersionID1,
			launcherVersion: domain.NewLauncherVersionWithoutQuestionnaire(
				launcherVersionID1,
				values.NewLauncherVersionName("2020.1.1"),
				now,
			),
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("game1"),
					values.NewGameDescription("description1"),
					now,
				),
				domain.NewGame(
					gameID2,
					values.NewGameName("game2"),
					values.NewGameDescription("description2"),
					now,
				),
			},
			expect: &openapi.VersionDetails{
				Id:        uuid.UUID(launcherVersionID1).String(),
				Name:      "2020.1.1",
				AnkeTo:    "",
				CreatedAt: now,
				Games: []openapi.GameMeta{
					{
						Id:   uuid.UUID(gameID1).String(),
						Name: "game1",
					},
					{
						Id:   uuid.UUID(gameID2).String(),
						Name: "game2",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			if testCase.executeAddGamesToLauncherVersion {
				mockLauncherVersionService.
					EXPECT().
					AddGamesToLauncherVersion(gomock.Any(), testCase.launcherVersionID, gomock.Any()).
					Return(testCase.launcherVersion, testCase.games, testCase.AddGamesToLauncherVersionErr)
			}

			launcherVersion, err := launcherVersionHandler.PostGameToVersion(testCase.strLauncherVersionID, testCase.apiGameIDs)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echo.HTTPError")
					}
				} else if testCase.err == nil {
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

			assert.Equal(t, testCase.expect.Id, launcherVersion.Id)
			assert.Equal(t, testCase.expect.Name, launcherVersion.Name)
			assert.Equal(t, testCase.expect.AnkeTo, launcherVersion.AnkeTo)
			assert.WithinDuration(t, testCase.expect.CreatedAt, launcherVersion.CreatedAt, time.Second)

			assert.Len(t, launcherVersion.Games, len(testCase.expect.Games))

			for i, expect := range testCase.expect.Games {
				assert.Equal(t, expect, launcherVersion.Games[i])
			}
		})
	}
}
