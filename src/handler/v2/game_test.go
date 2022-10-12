package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetGames(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	mockGameService := mock.NewMockGameV2(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description       string
		params            openapi.GetGamesParams
		executeGetGames   bool
		GetGamesErr       error
		sessionExist      bool
		authSession       *domain.OIDCSession
		executeGetMyGames bool
		GetMyGamesErr     error
		games             []*domain.Game
		apiGames          []*openapi.GameInfo
		gamesNumber       int
		isErr             bool
		err               error
		statusCode        int
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	now := time.Now()

	poTrue := true
	poFalse := false
	po0 := 0
	po1 := 1

	testCases := []test{
		{
			description: "特に問題ないので問題なし",
			params: openapi.GetGamesParams{
				All:    &poTrue,
				Limit:  &po0,
				Offset: &po0,
			},
			executeGetGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 1,
		},
		{
			description: "GetGamesがエラーなので500",
			params: openapi.GetGamesParams{
				All:    &poTrue,
				Limit:  &po0,
				Offset: &po0,
			},
			executeGetGames: true,
			GetGamesErr:     errors.New("test"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
		{
			description: "allがfalseなので問題なし",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po0,
				Offset: &po0,
			},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*domain.Game{domain.NewGame(
				gameID1,
				values.NewGameName("test1"),
				values.NewGameDescription("test1"),
				now,
			),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 1,
		},
		{
			description: "sessionが存在しないので500",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po0,
				Offset: &po0,
			},
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description: "authSessionが存在しないので500",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po0,
				Offset: &po0,
			},
			sessionExist: true,
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
		{
			description: "GetMyGamesがエラーなので500",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po0,
				Offset: &po0,
			},
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
			description: "ゲームが存在しなくても問題なし",
			params: openapi.GetGamesParams{
				All:    &poTrue,
				Limit:  &po0,
				Offset: &po0,
			},
			executeGetGames: true,
			games:           []*domain.Game{},
			apiGames:        []*openapi.GameInfo{},
		},
		{
			description: "ゲームが複数でも問題なし",
			params: openapi.GetGamesParams{
				All:    &poTrue,
				Limit:  &po0,
				Offset: &po0,
			},
			executeGetGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
				domain.NewGame(
					gameID2,
					values.NewGameName("test2"),
					values.NewGameDescription("test2"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					Id:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
		{
			description: "falseかつゲームが存在しなくても問題なし",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po0,
				Offset: &po0,
			},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games:             []*domain.Game{},
			apiGames:          []*openapi.GameInfo{},
		},
		{
			description: "falseかつゲームが複数でも問題なし",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po0,
				Offset: &po0,
			},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
				domain.NewGame(
					gameID2,
					values.NewGameName("test2"),
					values.NewGameDescription("test2"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
				{
					Id:          uuid.UUID(gameID2),
					Name:        "test2",
					Description: "test2",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
		{
			description: "Limitが設定されても問題なし",
			params: openapi.GetGamesParams{
				All:    &poTrue,
				Limit:  &po1,
				Offset: &po0,
			},
			executeGetGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
		{
			description: "Offsetが設定されても問題なし",
			params: openapi.GetGamesParams{
				All:    &poTrue,
				Limit:  &po0,
				Offset: &po1,
			},
			executeGetGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
		{
			description: "LimitとOffsetが設定されても問題なし",
			params: openapi.GetGamesParams{
				All:    &poTrue,
				Limit:  &po1,
				Offset: &po1,
			},
			executeGetGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
		{
			description: "allがfalseでLimitが設定されても問題なし",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po1,
				Offset: &po0,
			},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
		{
			description: "allがfalseでOffsetが設定されても問題なし",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po0,
				Offset: &po1,
			},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
		{
			description: "allがfalseでLimitとOffsetが設定されても問題なし",
			params: openapi.GetGamesParams{
				All:    &poFalse,
				Limit:  &po1,
				Offset: &po1,
			},
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			executeGetMyGames: true,
			games: []*domain.Game{
				domain.NewGame(
					gameID1,
					values.NewGameName("test1"),
					values.NewGameDescription("test1"),
					now,
				),
			},
			apiGames: []*openapi.GameInfo{
				{
					Id:          uuid.UUID(gameID1),
					Name:        "test1",
					Description: "test1",
					CreatedAt:   now,
				},
			},
			gamesNumber: 2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/games", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
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

				sess, err = session.Get(req)
				if err != nil {
					t.Fatal(err)
				}

				c.Set("session", sess)
			}

			if testCase.executeGetGames {
				mockGameService.
					EXPECT().
					GetGames(gomock.Any(), int(*testCase.params.Limit), int(*testCase.params.Offset)).
					Return(testCase.gamesNumber, testCase.games, testCase.GetGamesErr)
			}

			if testCase.executeGetMyGames {
				mockGameService.
					EXPECT().
					GetMyGames(gomock.Any(), gomock.Any(), int(*testCase.params.Limit), int(*testCase.params.Offset)).
					Return(testCase.gamesNumber, testCase.games, testCase.GetMyGamesErr)
			}

			err := gameHandler.GetGames(c, testCase.params)

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

			var response openapi.GetGamesResponse
			err = json.NewDecoder(rec.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			gamesNumber := response.Num
			games := response.Games

			assert.Equal(t, testCase.gamesNumber, gamesNumber)
			assert.Equal(t, len(testCase.apiGames), len(games))

			for i, game := range games {
				assert.Equal(t, testCase.apiGames[i].Id, game.Id)
				assert.Equal(t, testCase.apiGames[i].Name, game.Name)
				assert.Equal(t, testCase.apiGames[i].Description, game.Description)
				assert.WithinDuration(t, testCase.apiGames[i].CreatedAt, game.CreatedAt, time.Second)
			}
		})
	}
}

func TestPostGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGameV2(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description       string
		sessionExist      bool
		authSession       *domain.OIDCSession
		newGame           *openapi.PostGameJSONRequestBody
		executeCreateGame bool
		game              *service.GameInfoV2
		CreateGameErr     error
		apiGame           openapi.Game
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
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					now,
				),
				Owners: []*service.UserInfo{
					service.NewUserInfo(
						values.NewTrapMemberID(uuid.New()),
						"mazrean",
						values.TrapMemberStatusActive,
					),
					service.NewUserInfo(
						values.NewTrapMemberID(uuid.New()),
						"ikura-hamu",
						values.TrapMemberStatusActive,
					),
				},
				Maintainers: []*service.UserInfo{
					service.NewUserInfo(
						values.NewTrapMemberID(uuid.New()),
						"pikachu",
						values.TrapMemberStatusActive,
					),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
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
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
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
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
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
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"",
					now,
				),
				Owners: []*service.UserInfo{
					service.NewUserInfo(
						values.NewTrapMemberID(uuid.New()),
						"mazrean",
						values.TrapMemberStatusActive,
					),
					service.NewUserInfo(
						values.NewTrapMemberID(uuid.New()),
						"ikura-hamu",
						values.TrapMemberStatusActive,
					),
				},
				Maintainers: []*service.UserInfo{
					service.NewUserInfo(
						values.NewTrapMemberID(uuid.New()),
						"pikachu",
						values.TrapMemberStatusActive,
					),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
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
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
			},
			executeCreateGame: true,
			CreateGameErr:     errors.New("test"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
		{
			description:  "Ownersに重複があるので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
				Owners:      &[]openapi.UserName{"mazrean", "mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
			},
			executeCreateGame: true,
			CreateGameErr:     service.ErrOverlapInOwners,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:  "Maintainersに重複があるので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu", "pikachu"},
			},
			executeCreateGame: true,
			CreateGameErr:     service.ErrOverlapInMaintainers,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:  "Ownersに重複があるので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
				Owners:      &[]openapi.UserName{"mazrean", "mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
			},
			executeCreateGame: true,
			CreateGameErr:     service.ErrOverlapInOwners,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:  "OwnersとMaintainersに重複があるので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"mazrean"},
			},
			executeCreateGame: true,
			CreateGameErr:     service.ErrOverlapBetweenOwnersAndMaintainers,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()

			reqBody := new(bytes.Buffer)
			err = json.NewEncoder(reqBody).Encode(testCase.newGame)
			if err != nil {
				log.Printf("failed to create request body")
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/game", reqBody)
			req.Header.Set(echo.HeaderContentType, "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
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

				sess, err = session.Get(req)
				if err != nil {
					t.Fatal(err)
				}

				c.Set("session", sess)
			}

			if testCase.executeCreateGame {
				owners := make([]values.TraPMemberName, 0, len(*testCase.newGame.Owners))
				for _, owner := range *testCase.newGame.Owners {
					owners = append(owners, values.TraPMemberName(owner))
				}

				maintainers := make([]values.TraPMemberName, 0, len(*testCase.newGame.Maintainers))
				for _, maintainer := range *testCase.newGame.Maintainers {
					maintainers = append(maintainers, values.TraPMemberName(maintainer))
				}

				mockGameService.
					EXPECT().
					CreateGame(
						gomock.Any(),
						gomock.Any(),
						values.NewGameName(testCase.newGame.Name),
						values.NewGameDescription(testCase.newGame.Description),
						owners,
						maintainers).
					Return(testCase.game, testCase.CreateGameErr)
			}

			err := gameHandler.PostGame(c)

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

			var responseGame openapi.Game
			err = json.NewDecoder(rec.Body).Decode(&responseGame)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.apiGame.Id, responseGame.Id)
			assert.Equal(t, testCase.apiGame.Name, responseGame.Name)
			assert.Equal(t, testCase.apiGame.Description, responseGame.Description)
			assert.WithinDuration(t, testCase.apiGame.CreatedAt, responseGame.CreatedAt, time.Second)

			assert.Len(t, responseGame.Owners, len(testCase.apiGame.Owners))

			if *responseGame.Maintainers != nil {
				assert.Len(t, *responseGame.Maintainers, len(*testCase.apiGame.Maintainers))
			}

			for i, resOwner := range responseGame.Owners {
				assert.Equal(t, testCase.apiGame.Owners[i], resOwner)
			}
			for i, resMaintainer := range *responseGame.Maintainers {
				assert.Equal(t, (*testCase.apiGame.Maintainers)[i], resMaintainer)
			}
		})
	}
}

func TestDeleteGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGameV2(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description       string
		executeDeleteGame bool
		gameIDInPath      openapi.GameIDInPath
		gameID            values.GameID
		DeleteGameErr     error
		isErr             bool
		err               error
		statusCode        int
	}

	gameID := openapi.GameIDInPath(uuid.New())

	testCases := []test{
		{
			description:       "特に問題ないのでエラーなし",
			executeDeleteGame: true,
			gameIDInPath:      gameID,
			gameID:            values.GameID(gameID),
		},
		{
			description:       "ゲームが存在しないので400",
			executeDeleteGame: true,
			gameIDInPath:      gameID,
			gameID:            values.GameID(gameID),
			DeleteGameErr:     service.ErrNoGame,
			isErr:             true,
			statusCode:        http.StatusNotFound,
		},
		{
			description:       "DeleteGameがエラーなので500",
			executeDeleteGame: true,
			gameIDInPath:      gameID,
			gameID:            values.GameID(gameID),
			DeleteGameErr:     errors.New("test"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/game/%s", testCase.gameIDInPath), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeDeleteGame {
				mockGameService.
					EXPECT().
					DeleteGame(gomock.Any(), testCase.gameID).
					Return(testCase.DeleteGameErr)
			}

			err := gameHandler.DeleteGame(c, testCase.gameIDInPath)

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

func TestGetGame(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := common.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	mockGameService := mock.NewMockGameV2(ctrl)

	gameHandler := NewGame(session, mockGameService)

	type test struct {
		description    string
		sessionExist   bool
		authSession    *domain.OIDCSession
		gameIDInPath   openapi.GameIDInPath
		gameID         values.GameID
		executeGetGame bool
		game           *service.GameInfoV2
		GetGameErr     error
		apiGame        openapi.Game
		isErr          bool
		err            error
		statusCode     int
	}

	gameID := values.NewGameID()

	userID1 := values.NewTrapMemberID(uuid.New())
	userID2 := values.NewTrapMemberID(uuid.New())

	now := time.Now()

	testCases := []test{
		{
			description:  "特に問題ないのでエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			gameIDInPath:   openapi.GameID(gameID),
			gameID:         gameID,
			executeGetGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					now,
				),
				Owners: []*service.UserInfo{
					service.NewUserInfo(
						userID1,
						values.NewTrapMemberName("mazrean"),
						values.TrapMemberStatusActive,
					),
				},
				Maintainers: []*service.UserInfo{
					service.NewUserInfo(
						userID2,
						values.NewTrapMemberName("pikachu"),
						values.TrapMemberStatusActive,
					),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
			},
		},
		{
			description:  "Maintainersが空でもエラーなし",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			gameIDInPath:   openapi.GameID(gameID),
			gameID:         gameID,
			executeGetGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					now,
				),
				Owners: []*service.UserInfo{
					service.NewUserInfo(
						userID1,
						values.NewTrapMemberName("mazrean"),
						values.TrapMemberStatusActive,
					),
				},
				Maintainers: []*service.UserInfo{},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{},
			},
		},
		{
			description:  "ゲームが存在しないので404",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			gameIDInPath:   openapi.GameID(gameID),
			gameID:         gameID,
			executeGetGame: true,
			GetGameErr:     service.ErrNoGame,
			isErr:          true,
			statusCode:     http.StatusNotFound,
		},
		{
			description:  "GetGameがエラーなので500",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			gameIDInPath:   openapi.GameID(gameID),
			gameID:         gameID,
			executeGetGame: true,
			GetGameErr:     errors.New("error"),
			isErr:          true,
			statusCode:     http.StatusInternalServerError,
		},
		{
			description:  "セッションが存在しないので500",
			sessionExist: false,
			gameIDInPath: openapi.GameID(gameID),
			gameID:       gameID,
			GetGameErr:   errors.New("error"),
			isErr:        true,
			statusCode:   http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/game/%s", testCase.gameIDInPath), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.sessionExist {
				sess, err := session.New(req)
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

				sess, err = session.Get(req)
				if err != nil {
					t.Fatal(err)
				}

				c.Set("session", sess)
			}

			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Any(), testCase.gameID).
					Return(testCase.game, testCase.GetGameErr)
			}

			err := gameHandler.GetGame(c, testCase.gameIDInPath)

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

			var responseGame openapi.Game
			err = json.NewDecoder(rec.Body).Decode(&responseGame)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			assert.Equal(t, testCase.apiGame.Name, responseGame.Name)
			assert.Equal(t, testCase.apiGame.Id, responseGame.Id)
			assert.Equal(t, testCase.apiGame.Description, responseGame.Description)
			assert.WithinDuration(t, testCase.apiGame.CreatedAt, responseGame.CreatedAt, time.Second)

			assert.Len(t, testCase.apiGame.Owners, len(responseGame.Owners))
			for i, apiOwner := range testCase.apiGame.Owners {
				assert.Equal(t, apiOwner, responseGame.Owners[i])
			}

			if len(*responseGame.Maintainers) != 0 {
				assert.Len(t, *testCase.apiGame.Maintainers, len(*responseGame.Maintainers))
				for i, apiMaintainer := range *testCase.apiGame.Maintainers {
					assert.Equal(t, apiMaintainer, []string(*responseGame.Maintainers)[i])
				}
			}
		})
	}
}
