package v1

import (
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestPostGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	session := NewSession("key", "secret")
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

			game, err := gameHandler.PostGame(testCase.newGame, c)

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

	session := NewSession("key", "secret")
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
			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Any()).
					Return(testCase.game, testCase.GetGameErr)
			}

			game, err := gameHandler.GetGame(testCase.strGameID)

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
