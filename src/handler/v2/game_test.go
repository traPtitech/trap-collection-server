package v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
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

	now := time.Now()

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameName1 := values.NewGameName("test1")
	gameName2 := values.NewGameName("test2")
	gameName3 := values.NewGameName("test3")
	gameDescription1 := values.NewGameDescription("test1")
	gameDescription2 := values.NewGameDescription("test2")
	gameDescription3 := values.NewGameDescription("test3")
	game1 := domain.NewGame(gameID1, gameName1, gameDescription1, values.GameVisibilityTypePublic, now.Add(-time.Hour))
	game2 := domain.NewGame(gameID2, gameName2, gameDescription2, values.GameVisibilityTypeLimited, now.Add(-time.Hour*2))
	game3 := domain.NewGame(gameID3, gameName3, gameDescription3, values.GameVisibilityTypePrivate, now.Add(-time.Hour*3))

	gameGenreID1 := values.NewGameGenreID()
	gameGenreID2 := values.NewGameGenreID()
	gameGenreID3 := values.NewGameGenreID()
	gameGenreName1 := values.NewGameGenreName("3D")
	gameGenreName2 := values.NewGameGenreName("2D")
	gameGenreName3 := values.NewGameGenreName("VR")
	gameGenre1 := domain.NewGameGenre(gameGenreID1, gameGenreName1, now.Add(-time.Hour))
	gameGenre2 := domain.NewGameGenre(gameGenreID2, gameGenreName2, now.Add(-time.Hour*2))
	gameGenre3 := domain.NewGameGenre(gameGenreID3, gameGenreName3, now.Add(-time.Hour*3))

	valueTrue := true
	valueFalse := false
	value0 := 0
	value1 := 1
	value3 := 3
	sortTypeCreatedAt := openapi.CreatedAt
	sortTypeLatestVersion := openapi.LatestVersion
	gameNameStr := "test"

	type test struct {
		params            openapi.GetGamesParams
		sessionExist      bool
		authSession       *domain.OIDCSession
		executeGetGames   bool
		GetGamesErr       error
		executeGetMyGames bool
		GetMyGamesErr     error
		games             []*domain.GameWithGenres
		gamesNumber       int
		apiGames          openapi.GetGamesResponse
		isErr             bool
		err               error
		statusCode        int
	}

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
				},
				Num: 2,
			},
		},
		"ゲームが複数あっても問題なし": {
			params: openapi.GetGamesParams{
				Limit:  &value3,
				Offset: &value0,
				All:    &valueTrue,
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			games: []*domain.GameWithGenres{
				domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1}),
				domain.NewGameWithGenres(game2, []*domain.GameGenre{gameGenre2}),
				domain.NewGameWithGenres(game3, []*domain.GameGenre{gameGenre3}),
			},
			gamesNumber: 3,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
					{
						Id:          uuid.UUID(gameID2),
						Name:        string(gameName2),
						Description: string(gameDescription2),
						Visibility:  openapi.Limited,
						CreatedAt:   now.Add(-time.Hour * 2),
						Genres:      &[]string{string(gameGenreName2)},
					},
					{
						Id:          uuid.UUID(gameID3),
						Name:        string(gameName3),
						Description: string(gameDescription3),
						Visibility:  openapi.Private,
						CreatedAt:   now.Add(-time.Hour * 3),
						Genres:      &[]string{string(gameGenreName3)},
					},
				},
				Num: 3,
			},
		},
		"ジャンルが複数あってもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1, gameGenre2})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1), string(gameGenreName2)},
					},
				},
				Num: 2,
			},
		},
		"authSessionが無くてもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
			},
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
				},
				Num: 2,
			},
		},
		"allがfalseでもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueFalse,
			},
			sessionExist:      true,
			authSession:       domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetMyGames: true,
			games:             []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			gamesNumber:       2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
				},
				Num: 2,
			},
		},
		"allがfalseでauthSessionが無くてもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueFalse,
			},
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
				},
				Num: 2,
			},
		},
		"limitとoffsetとallがnilでもエラー無し": {
			params:          openapi.GetGamesParams{},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			apiGames:        openapi.GetGamesResponse{},
		},
		"sortがCreatedAtでもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
				Sort:   &sortTypeCreatedAt,
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
				},
				Num: 2,
			},
		},
		"sortがLatestVersionでもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
				Sort:   &sortTypeLatestVersion,
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
				},
				Num: 2,
			},
		},
		"sortが変な値なので400": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
				Sort:   new(openapi.GetGamesParamsSort),
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		"ジャンルの指定があってもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
				Genre:  &[]uuid.UUID{uuid.UUID(gameGenreID1), uuid.UUID(gameGenreID2)},
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1, gameGenre2})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1), string(gameGenreName2)},
					},
				},
				Num: 2,
			},
		},
		"ゲーム名の指定があってもエラー無し": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
				Name:   &gameNameStr,
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			games:           []*domain.GameWithGenres{domain.NewGameWithGenres(game1, []*domain.GameGenre{gameGenre1})},
			gamesNumber:     2,
			apiGames: openapi.GetGamesResponse{
				Games: []openapi.GameInfoWithGenres{
					{
						Id:          uuid.UUID(gameID1),
						Name:        string(gameName1),
						Description: string(gameDescription1),
						Visibility:  openapi.Public,
						CreatedAt:   now.Add(-time.Hour),
						Genres:      &[]string{string(gameGenreName1)},
					},
				},
				Num: 2,
			},
		},
		"GetGamesがエラーなのでエラー": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueTrue,
			},
			sessionExist:    true,
			authSession:     domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetGames: true,
			GetGamesErr:     errors.New("GetGames error"),
			isErr:           true,
			statusCode:      http.StatusInternalServerError,
		},
		"GetMyGamesがエラーなのでエラー": {
			params: openapi.GetGamesParams{
				Limit:  &value1,
				Offset: &value0,
				All:    &valueFalse,
			},
			sessionExist:      true,
			authSession:       domain.NewOIDCSession("token", now.Add(time.Hour)),
			executeGetMyGames: true,
			GetMyGamesErr:     errors.New("GetGames error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
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
				limit, offset := 0, 0
				if testCase.params.Limit != nil {
					limit = *testCase.params.Limit
				}
				if testCase.params.Offset != nil {
					offset = *testCase.params.Offset
				}
				mockGameService.
					EXPECT().
					GetGames(
						gomock.Any(), limit, offset,
						gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(testCase.gamesNumber, testCase.games, testCase.GetGamesErr)
			}

			if testCase.executeGetMyGames {
				mockGameService.
					EXPECT().
					GetMyGames(
						gomock.Any(), gomock.Not(gomock.Nil()),
						int(*testCase.params.Limit), int(*testCase.params.Offset), gomock.Any(),
						gomock.InAnyOrder([]values.GameVisibility{values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate, values.GameVisibilityTypePublic}),
						gomock.Any(), gomock.Any()).
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
			assert.Equal(t, len(testCase.apiGames.Games), len(games))

			for i, game := range games {
				assert.Equal(t, testCase.apiGames.Games[i].Id, game.Id)
				assert.Equal(t, testCase.apiGames.Games[i].Name, game.Name)
				assert.Equal(t, testCase.apiGames.Games[i].Description, game.Description)
				assert.WithinDuration(t, testCase.apiGames.Games[i].CreatedAt, game.CreatedAt, time.Second)

				assert.Len(t, *games[i].Genres, len(*testCase.apiGames.Games[i].Genres))
				for j := range *testCase.apiGames.Games[i].Genres {
					assert.Equal(t, (*testCase.apiGames.Games[i].Genres)[j], (*games[i].Genres)[j])
				}
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
		isBadRequestBody  bool
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

	var (
		visibilityPublic  openapi.GameVisibility = openapi.Public
		visibilityLimited openapi.GameVisibility = openapi.Limited
		visibilityPrivate openapi.GameVisibility = openapi.Private
		invalidVisibility openapi.GameVisibility = openapi.GameVisibility("invalid")
	)

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
				Genres:      &[]openapi.GameGenreName{"3D"},
				Visibility:  &visibilityPublic,
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					values.GameVisibilityTypePublic,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(values.GameGenreIDFromUUID(uuid.New()), "3D", now),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Public,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
		},
		{
			description:  "セッションがないので401",
			sessionExist: false,
			isErr:        true,
			statusCode:   http.StatusUnauthorized,
		},
		{
			description:  "authSessionがないので401",
			sessionExist: true,
			isErr:        true,
			statusCode:   http.StatusUnauthorized,
		},
		{
			description:      "リクエストボディが正しくないので400",
			sessionExist:     true,
			isBadRequestBody: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			isErr:      true,
			statusCode: http.StatusBadRequest,
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"",
					values.GameVisibilityTypePublic,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(values.NewGameGenreID(), "3D", now),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Public,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
		},
		{
			description:  "Genresが空でもエラー無し",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "すごいゲーム",
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  &visibilityPublic,
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"すごいゲーム",
					values.GameVisibilityTypePublic,
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
				Genres: []*domain.GameGenre{},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "すごいゲーム",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Public,
				Genres:      nil,
			},
		},
		{
			description:  "ジャンル名が長すぎるので400",
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{strings.Repeat("あ", 33)},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:  "ジャンル名が0文字なので400",
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{""},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description:  "ジャンルが複数あっても問題なし",
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
				Genres:      &[]openapi.GameGenreName{"3D", "2D"},
				Visibility:  &visibilityPublic,
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					values.GameVisibilityTypePublic,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(values.NewGameGenreID(), "3D", now),
					domain.NewGameGenre(values.NewGameGenreID(), "2D", now),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Public,
				Genres:      &[]openapi.GameGenreName{"3D", "2D"},
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
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
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
			executeCreateGame: true,
			CreateGameErr:     service.ErrOverlapBetweenOwnersAndMaintainers,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:  "ジャンルに重複があるので400",
			sessionExist: true,
			authSession: domain.NewOIDCSession(
				"accessToken",
				time.Now().Add(time.Hour),
			),
			newGame: &openapi.NewGame{
				Name:        "test",
				Description: "test",
				Owners:      &[]openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"ikura-hamu"},
				Visibility:  &visibilityPublic,
				Genres:      &[]openapi.GameGenreName{"3D", "3D"},
			},
			executeCreateGame: true,
			CreateGameErr:     service.ErrDuplicateGameGenre,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:  "visibilityがlimitedでも問題なし",
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
				Genres:      &[]openapi.GameGenreName{"3D"},
				Visibility:  &visibilityLimited,
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					values.GameVisibilityTypeLimited,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(values.GameGenreIDFromUUID(uuid.New()), "3D", now),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Limited,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
		},
		{
			description:  "visibilityがprivateでも問題なし",
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
				Genres:      &[]openapi.GameGenreName{"3D"},
				Visibility:  &visibilityPrivate,
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					values.GameVisibilityTypePrivate,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(values.GameGenreIDFromUUID(uuid.New()), "3D", now),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Private,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
		},
		{
			description:  "visibilityがnilでもprivateになる",
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
				Genres:      &[]openapi.GameGenreName{"3D"},
				Visibility:  nil,
			},
			executeCreateGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					values.GameVisibilityTypePrivate,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(values.GameGenreIDFromUUID(uuid.New()), "3D", now),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean", "ikura-hamu"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Private,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
		},
		{
			description:  "visibilityの値が正しくないので400",
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
				Visibility:  &invalidVisibility,
				Genres:      &[]openapi.GameGenreName{"3D"},
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()

			reqBody := new(bytes.Buffer)
			if !testCase.isBadRequestBody {
				err = json.NewEncoder(reqBody).Encode(testCase.newGame)
				if err != nil {
					log.Printf("failed to create request body")
					t.Fatal(err)
				}
			} else {
				reqBody = bytes.NewBufferString("bad requset body")
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
						gomock.Any(),
						owners,
						maintainers,
						gomock.Any()).
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

			if responseGame.Genres != nil {
				assert.Len(t, *responseGame.Genres, len(*testCase.apiGame.Genres))
				for i, resGenre := range *responseGame.Genres {
					assert.Equal(t, (*testCase.apiGame.Genres)[i], resGenre)
				}
			} else {
				assert.Nil(t, testCase.apiGame.Genres)
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
			description:       "ゲームが存在しないので404",
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
	genreID := values.NewGameGenreID()

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
					values.GameVisibilityTypeLimited,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(genreID, "test", now.Add(-time.Hour)),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Visibility:  openapi.Limited,
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
					values.GameVisibilityTypeLimited,
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
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(genreID, "test", now.Add(-time.Hour)),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{},
				Genres:      &[]openapi.GameGenreName{"test"},
				Visibility:  openapi.Limited,
			},
		},
		{
			description:  "Genresが空でもエラーなし",
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
					values.GameVisibilityTypeLimited,
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
				Genres: []*domain.GameGenre{},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Genres:      &[]openapi.GameGenreName{},
				Visibility:  openapi.Limited,
			},
		},
		{
			description:  "visibilityがprivateでも問題なし",
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
					values.GameVisibilityTypePrivate,
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
				Genres: []*domain.GameGenre{},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Genres:      &[]openapi.GameGenreName{},
				Visibility:  openapi.Private,
			},
		},
		{
			description:  "visibilityがpublicでもエラーなし",
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
					values.GameVisibilityTypePublic,
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
				Genres: []*domain.GameGenre{},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{"mazrean"},
				Maintainers: &[]openapi.UserName{"pikachu"},
				Genres:      &[]openapi.GameGenreName{},
				Visibility:  openapi.Public,
			},
		},
		{
			description:  "visibilityが不正なので500",
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
					100,
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
				Genres: []*domain.GameGenre{},
			},
			isErr:      true,
			statusCode: http.StatusInternalServerError,
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
			description:    "authSessionが存在しないので管理者は空配列",
			sessionExist:   true,
			gameIDInPath:   openapi.GameID(gameID),
			gameID:         gameID,
			executeGetGame: true,
			game: &service.GameInfoV2{
				Game: domain.NewGame(
					gameID,
					"test",
					"test",
					values.GameVisibilityTypeLimited,
					now,
				),
				Owners:      []*service.UserInfo{},
				Maintainers: []*service.UserInfo{},
				Genres: []*domain.GameGenre{
					domain.NewGameGenre(genreID, "test", now.Add(-time.Hour)),
				},
			},
			apiGame: openapi.Game{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
				Owners:      []openapi.UserName{},
				Maintainers: &[]openapi.UserName{},
				Genres:      &[]openapi.GameGenreName{"test"},
				Visibility:  openapi.Limited,
			},
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
			assert.Equal(t, testCase.apiGame.Visibility, responseGame.Visibility)
			assert.WithinDuration(t, testCase.apiGame.CreatedAt, responseGame.CreatedAt, time.Second)

			assert.Len(t, testCase.apiGame.Owners, len(responseGame.Owners))
			for i, apiOwner := range testCase.apiGame.Owners {
				assert.Equal(t, apiOwner, responseGame.Owners[i])
			}

			assert.Len(t, *testCase.apiGame.Maintainers, len(*responseGame.Maintainers))
			for i, apiMaintainer := range *testCase.apiGame.Maintainers {
				assert.Equal(t, apiMaintainer, []string(*responseGame.Maintainers)[i])
			}
		})
	}
}

func TestPatchGame(t *testing.T) {
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
		isBadRequestBody  bool
		gameID            values.GameID
		newGame           *openapi.PatchGameJSONRequestBody
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
			gameID:      gameID,
			newGame: &openapi.PatchGameJSONRequestBody{
				Name:        "test",
				Description: "test",
			},
			executeUpdateGame: true,
			game: domain.NewGame(
				gameID,
				values.NewGameName("test"),
				values.NewGameDescription("test"),
				values.GameVisibilityTypeLimited,
				now,
			),
			apiGame: openapi.GameInfo{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "test",
				CreatedAt:   now,
			},
		},
		{
			description:      "リクエストボディが正しくないので400",
			isBadRequestBody: true,
			isErr:            true,
			statusCode:       http.StatusBadRequest,
		},
		{
			description: "名前が空なので400",
			gameID:      gameID,
			newGame: &openapi.PatchGameJSONRequestBody{
				Name:        "",
				Description: "test",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "名前が長すぎるので400",
			gameID:      gameID,
			newGame: &openapi.PatchGameJSONRequestBody{
				Name:        "012345678901234567890123456789012",
				Description: "test",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		{
			description: "説明が空文字でもエラーなし",
			gameID:      gameID,
			newGame: &openapi.PatchGameJSONRequestBody{
				Name:        "test",
				Description: "",
			},
			executeUpdateGame: true,
			game: domain.NewGame(
				gameID,
				values.NewGameName("test"),
				values.NewGameDescription(""),
				values.GameVisibilityTypeLimited,
				now,
			),
			apiGame: openapi.GameInfo{
				Id:          uuid.UUID(gameID),
				Name:        "test",
				Description: "",
				CreatedAt:   now,
			},
		},
		{
			description: "ゲームが存在しないので404",
			gameID:      gameID,
			newGame: &openapi.PatchGameJSONRequestBody{
				Name:        "test",
				Description: "test",
			},
			executeUpdateGame: true,
			UpdateGameErr:     service.ErrNoGame,
			isErr:             true,
			statusCode:        http.StatusNotFound,
		},
		{
			description: "UpdateGameがエラーなので500",
			gameID:      gameID,
			newGame: &openapi.PatchGameJSONRequestBody{
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

			reqBody := new(bytes.Buffer)
			if !testCase.isBadRequestBody {
				err = json.NewEncoder(reqBody).Encode(testCase.newGame)
				if err != nil {
					log.Printf("failed to create request body")
					t.Fatal(err)
				}
			} else {
				reqBody = bytes.NewBufferString("bad request body")
			}

			err = json.NewEncoder(reqBody).Encode(testCase.newGame)
			if err != nil {
				log.Printf("failed to create request body")
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/game/%s", uuid.UUID(gameID)), reqBody)
			req.Header.Set(echo.HeaderContentType, "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeUpdateGame {
				mockGameService.
					EXPECT().
					UpdateGame(gomock.Any(), gomock.Any(), values.NewGameName(testCase.newGame.Name), values.NewGameDescription(testCase.newGame.Description)).
					Return(testCase.game, testCase.UpdateGameErr)
			}

			err := gameHandler.PatchGame(c, openapi.GameID(gameID))

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

			var responseGame openapi.GameInfo
			err = json.NewDecoder(rec.Body).Decode(&responseGame)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			assert.Equal(t, testCase.apiGame.Name, responseGame.Name)
			assert.Equal(t, testCase.apiGame.Id, responseGame.Id)
			assert.Equal(t, testCase.apiGame.Description, responseGame.Description)
			assert.WithinDuration(t, testCase.apiGame.CreatedAt, responseGame.CreatedAt, time.Second)
		})
	}
}
