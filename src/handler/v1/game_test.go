package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestPostGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandlerV1(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	session, err := NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGame(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description       string
		sessionExist      bool
		authSession       *domain.OIDCSession
		newGame           *openapi.NewGame
		executeCreateGame bool
		game              *domain.Game
		CreateGameErr     error
		apiGame           openapi.GameInfo
		isErr             bool
		err               error
		statusCode        int
	}

	gameID := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description:  "特に問題ないのでエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
			},
			executeCreateGame: true,
			game: domain.NewGame(
				gameID,
				values.NewGameName("test"),
				values.NewGameDescription("test"),
				now,
			),
			apiGame: openapi.GameInfo{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
			},
		},
		{
			description:  "セッションがないので500",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "authSessionがないので500",
			sessionExist: true,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "名前が空なので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "",
				Description: "test",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:  "名前が長すぎるので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "012345678901234567890123456789012",
				Description: "test",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:  "説明が空文字でもエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "",
			},
			executeCreateGame: true,
			game: domain.NewGame(
				gameID,
				values.NewGameName("test"),
				values.NewGameDescription(""),
				now,
			),
			apiGame: openapi.GameInfo{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "",
				CreatedAt:   now,
			},
		},
		{
			description:  "CreateGameがエラーなので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
			},
			executeCreateGame: true,
			CreateGameErr:     errors.New("test"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/game", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSession != nil {
					sess.Values[accessTokenSessionKey] = string(testCase.authSession.GetAccessToken())
					sess.Values[expiresAtSessionKey] = testCase.authSession.GetExpiresAt()
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)

				sess, err = session.store.Get(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				c.Set(sessionContextKey, sess)
			}

			if testCase.executeCreateGame {
				mockGameService.
					EXPECT().
					CreateGame(gomock.Any(), gomock.Any(), values.NewGameName(testCase.newGame.Name), values.NewGameDescription(testCase.newGame.Description)).
					Return(testCase.game, testCase.CreateGameErr)
			}

			game, err := gameHandler.PostGame(c, testCase.newGame)

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

			assert.Equal(t, testCase.apiGame, *game)
		})
	}
}

func TestGetGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandlerV1(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	session, err := NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGame(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description    string
		strGameID      string
		executeGetGame bool
		game           *service.GameInfo
		GetGameErr     error
		apiGame        openapi.Game
		isErr          bool
		err            error
		statusCode     int
	}

	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()

	now := time.Now()

	testCases := []test{
		{
			description:    "特に問題ないのでエラーなし",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			game: &service.GameInfo{
				Game: domain.NewGame(
					gameID,
					values.NewGameName("test"),
					values.NewGameDescription("test"),
					now,
				),
				LatestVersion: domain.NewGameVersion(
					gameVersionID,
					values.NewGameVersionName("test"),
					values.NewGameVersionDescription("test"),
					now,
				),
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Version: &openapi.GameVersion{
					Id:          uuid.UUID(gameVersionID).String(),
					Name:        "test",
					Description: "test",
					CreatedAt:   now,
				},
			},
		},
		{
			description: "gameIDがuuidでないので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:    "ゲームが存在しないので400",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			GetGameErr:     service.ErrNoGame,
			isErr:          true,
			statusCode:     http.StatusBadRequest,
		},
		{
			description:    "GetGameがエラーなので500",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			GetGameErr:     errors.New("error"),
			isErr:          true,
			statusCode:     http.StatusInternalServerError,
		},
		{
			description:    "gameVersionがnilでもエラーなし",
			strGameID:      uuid.UUID(gameID).String(),
			executeGetGame: true,
			game: &service.GameInfo{
				Game: domain.NewGame(
					gameID,
					values.NewGameName("test"),
					values.NewGameDescription("test"),
					now,
				),
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/game/%s", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Any()).
					Return(testCase.game, testCase.GetGameErr)
			}

			game, err := gameHandler.GetGame(c, testCase.strGameID)

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

			assert.Equal(t, testCase.apiGame, *game)
		})
	}
}

func TestPutGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandlerV1(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	session, err := NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGame(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description       string
		strGameID         string
		newGame           *openapi.NewGame
		executeUpdateGame bool
		game              *domain.Game
		UpdateGameErr     error
		apiGame           openapi.GameInfo
		isErr             bool
		err               error
		statusCode        int
	}

	gameID := values.NewGameID()

	now := time.Now()

	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			strGameID:   uuid.UUID(gameID).String(),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
			},
			executeUpdateGame: true,
			game: domain.NewGame(
				gameID,
				values.NewGameName("test"),
				values.NewGameDescription("test"),
				now,
			),
			apiGame: openapi.GameInfo{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
			},
		},
		{
			description: "gameIDがuuidでないので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description: "名前が空なので400",
			strGameID:   uuid.UUID(gameID).String(),
			newGame: &openapi.NewGame{
				Name:        "",
				Description: "test",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "名前が長すぎるので400",
			strGameID:   uuid.UUID(gameID).String(),
			newGame: &openapi.NewGame{
				Name:        "012345678901234567890123456789012",
				Description: "test",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "説明が空文字でもエラーなし",
			strGameID:   uuid.UUID(gameID).String(),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "",
			},
			executeUpdateGame: true,
			game: domain.NewGame(
				gameID,
				values.NewGameName("test"),
				values.NewGameDescription(""),
				now,
			),
			apiGame: openapi.GameInfo{
				Id:          uuid.UUID(gameID).String(),
				Name:        "test",
				Description: "",
				CreatedAt:   now,
			},
		},
		{
			description: "ゲームが存在しないので400",
			strGameID:   uuid.UUID(gameID).String(),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
			},
			executeUpdateGame: true,
			UpdateGameErr:     service.ErrNoGame,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description: "UpdateGameがエラーなので500",
			strGameID:   uuid.UUID(gameID).String(),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
			},
			executeUpdateGame: true,
			UpdateGameErr:     errors.New("test"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/game/%s", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeUpdateGame {
				mockGameService.
					EXPECT().
					UpdateGame(gomock.Any(), gomock.Any(), values.NewGameName(testCase.newGame.Name), values.NewGameDescription(testCase.newGame.Description)).
					Return(testCase.game, testCase.UpdateGameErr)
			}

			game, err := gameHandler.PutGame(c, testCase.strGameID, testCase.newGame)

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

			assert.Equal(t, testCase.apiGame, *game)
		})
	}
}

func TestGetGames(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandlerV1(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	session, err := NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGame(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description       string
		strAll            string
		executeGetGames   bool
		GetGamesErr       error
		sessionExist      bool
		authSession       *domain.OIDCSession
		executeGetMyGames bool
		GetMyGamesErr     error
		games             []*service.GameInfo
		apiGames          []*openapi.Game
		isErr             bool
		err               error
		statusCode        int
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	gameVersionID1 := values.NewGameVersionID()
	gameVersionID2 := values.NewGameVersionID()

	now := time.Now()

	testCases := []test{
		{
			description:     "特に問題ないので問題なし",
			strAll:          "true",
			executeGetGames: true,
			games: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("test1"),
						values.NewGameDescription("test1"),
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("test1"),
						values.NewGameVersionDescription("test1"),
						now,
					),
				},
			},
			apiGames: []*openapi.Game{
				{
					Id:          uuid.UUID(gameID1).String(),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					Version: &openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1).String(),
						Name:        "test1",
						Description: "test1",
						CreatedAt:   now,
					},
				},
			},
		},
		{
			description:     "GetGamesがエラーなので500",
			strAll:          "true",
			executeGetGames: true,
			GetGamesErr:     errors.New("test"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
		{
			description: "allが誤っているので400",
			strAll:      "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:  "allがfalseなので問題なし",
			strAll:       "false",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("test1"),
						values.NewGameDescription("test1"),
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("test1"),
						values.NewGameVersionDescription("test1"),
						now,
					),
				},
			},
			apiGames: []*openapi.Game{
				{
					Id:          uuid.UUID(gameID1).String(),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					Version: &openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1).String(),
						Name:        "test1",
						Description: "test1",
						CreatedAt:   now,
					},
				},
			},
		},
		{
			description:  "allが空文字でも問題なし",
			strAll:       "",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("test1"),
						values.NewGameDescription("test1"),
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("test1"),
						values.NewGameVersionDescription("test1"),
						now,
					),
				},
			},
			apiGames: []*openapi.Game{
				{
					Id:          uuid.UUID(gameID1).String(),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					Version: &openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1).String(),
						Name:        "test1",
						Description: "test1",
						CreatedAt:   now,
					},
				},
			},
		},
		{
			description:  "sessionが存在しないので500",
			strAll:       "false",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "authSessionが存在しないので500",
			strAll:       "false",
			sessionExist: true,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description:  "GetMyGamesがエラーなので500",
			strAll:       "false",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			GetMyGamesErr:     errors.New("test"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
		{
			description:     "ゲームが存在しなくても問題なし",
			strAll:          "true",
			executeGetGames: true,
			games:           []*service.GameInfo{},
			apiGames:        []*openapi.Game{},
		},
		{
			description:     "ゲームが複数でも問題なし",
			strAll:          "true",
			executeGetGames: true,
			games: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("test1"),
						values.NewGameDescription("test1"),
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("test1"),
						values.NewGameVersionDescription("test1"),
						now,
					),
				},
				{
					Game: domain.NewGame(
						gameID2,
						values.NewGameName("test2"),
						values.NewGameDescription("test2"),
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID2,
						values.NewGameVersionName("test2"),
						values.NewGameVersionDescription("test2"),
						now,
					),
				},
			},
			apiGames: []*openapi.Game{
				{
					Id:          uuid.UUID(gameID1).String(),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					Version: &openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1).String(),
						Name:        "test1",
						Description: "test1",
						CreatedAt:   now,
					},
				},
				{
					Id:          uuid.UUID(gameID2).String(),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now,
					Version: &openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID2).String(),
						Name:        "test2",
						Description: "test2",
						CreatedAt:   now,
					},
				},
			},
		},
		{
			description:     "versionが存在しなくても問題なし",
			strAll:          "true",
			executeGetGames: true,
			games: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("test1"),
						values.NewGameDescription("test1"),
						now,
					),
				},
			},
			apiGames: []*openapi.Game{
				{
					Id:          uuid.UUID(gameID1).String(),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
		},
		{
			description:  "falseかつゲームが存在しなくても問題なし",
			strAll:       "false",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games:             []*service.GameInfo{},
			apiGames:          []*openapi.Game{},
		},
		{
			description:  "falseかつゲームが複数でも問題なし",
			strAll:       "false",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("test1"),
						values.NewGameDescription("test1"),
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID1,
						values.NewGameVersionName("test1"),
						values.NewGameVersionDescription("test1"),
						now,
					),
				},
				{
					Game: domain.NewGame(
						gameID2,
						values.NewGameName("test2"),
						values.NewGameDescription("test2"),
						now,
					),
					LatestVersion: domain.NewGameVersion(
						gameVersionID2,
						values.NewGameVersionName("test2"),
						values.NewGameVersionDescription("test2"),
						now,
					),
				},
			},
			apiGames: []*openapi.Game{
				{
					Id:          uuid.UUID(gameID1).String(),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
					Version: &openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID1).String(),
						Name:        "test1",
						Description: "test1",
						CreatedAt:   now,
					},
				},
				{
					Id:          uuid.UUID(gameID2).String(),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now,
					Version: &openapi.GameVersion{
						Id:          uuid.UUID(gameVersionID2).String(),
						Name:        "test2",
						Description: "test2",
						CreatedAt:   now,
					},
				},
			},
		},
		{
			description:  "falseかつversionが存在しなくても問題なし",
			strAll:       "false",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*service.GameInfo{
				{
					Game: domain.NewGame(
						gameID1,
						values.NewGameName("test1"),
						values.NewGameDescription("test1"),
						now,
					),
				},
			},
			apiGames: []*openapi.Game{
				{
					Id:          uuid.UUID(gameID1).String(),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/games", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.store.New(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				if testCase.authSession != nil {
					sess.Values[accessTokenSessionKey] = string(testCase.authSession.GetAccessToken())
					sess.Values[expiresAtSessionKey] = testCase.authSession.GetExpiresAt()
				}

				err = sess.Save(req, rec)
				if err != nil {
					t.Fatalf("failed to save session: %v", err)
				}

				setCookieHeader(c)

				sess, err = session.store.Get(req, session.key)
				if err != nil {
					t.Fatal(err)
				}

				c.Set(sessionContextKey, sess)
			}

			if testCase.executeGetGames {
				mockGameService.
					EXPECT().
					GetGames(gomock.Any()).
					Return(testCase.games, testCase.GetGamesErr)
			}

			if testCase.executeGetMyGames {
				mockGameService.
					EXPECT().
					GetMyGames(gomock.Any(), gomock.Any()).
					Return(testCase.games, testCase.GetMyGamesErr)
			}

			games, err := gameHandler.GetGames(c, testCase.strAll)

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

			assert.Len(t, games, len(testCase.apiGames))

			for i, game := range games {
				assert.Equal(t, testCase.apiGames[i].Id, game.Id)
				assert.Equal(t, testCase.apiGames[i].Name, game.Name)
				assert.Equal(t, testCase.apiGames[i].Description, game.Description)
				assert.Equal(t, testCase.apiGames[i].CreatedAt, game.CreatedAt)

				if testCase.apiGames[i].Version != nil {
					assert.Equal(t, testCase.apiGames[i].Version.Id, game.Version.Id)
					assert.Equal(t, testCase.apiGames[i].Version.Name, game.Version.Name)
					assert.Equal(t, testCase.apiGames[i].Version.Description, game.Version.Description)
					assert.Equal(t, testCase.apiGames[i].Version.CreatedAt, game.Version.CreatedAt)
				} else {
					assert.Nil(t, game.Version)
				}
			}
		})
	}
}

func TestDeleteGames(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandlerV1(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	session, err := NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGame(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description       string
		strGameID         string
		executeDeleteGame bool
		gameID            values.GameID
		DeleteGameErr     error
		isErr             bool
		err               error
		statusCode        int
	}

	gameID := values.NewGameID()

	testCases := []test{
		{
			description:       "特に問題ないのでエラーなし",
			strGameID:         uuid.UUID(gameID).String(),
			executeDeleteGame: true,
			gameID:            gameID,
		},
		{
			description: "gameIDが不正なので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:       "ゲームが存在しないので400",
			strGameID:         uuid.UUID(gameID).String(),
			executeDeleteGame: true,
			gameID:            gameID,
			DeleteGameErr:     service.ErrNoGame,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "DeleteGameがエラーなので500",
			strGameID:         uuid.UUID(gameID).String(),
			executeDeleteGame: true,
			gameID:            gameID,
			DeleteGameErr:     errors.New("test"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/game/%s", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeDeleteGame {
				mockGameService.
					EXPECT().
					DeleteGame(gomock.Any(), testCase.gameID).
					Return(testCase.DeleteGameErr)
			}

			err := gameHandler.DeleteGames(c, testCase.strGameID)

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
		})
	}
}
