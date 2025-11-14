package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mockConfig "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/session"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestDeleteGameGenre(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreService := mock.NewMockGameGenre(ctrl)

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)

	sess, err := session.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	gameGenre := NewGameGenre(mockGameGenreService, mock.NewMockGameV2(ctrl), session)

	type test struct {
		genreID            openapi.GameGenreIDInPath
		DeleteGameGenreErr error
		isErr              bool
		expectedErr        error
		statusCode         int
	}

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			genreID: uuid.New(),
		},
		"存在しないジャンルIDなので404": {
			genreID:            uuid.New(),
			DeleteGameGenreErr: service.ErrNoGameGenre,
			isErr:              true,
			statusCode:         http.StatusNotFound,
		},
		"DeleteGameGenreがエラーなのでエラー": {
			genreID:            uuid.New(),
			DeleteGameGenreErr: errors.New("error"),
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			mockGameGenreService.
				EXPECT().
				DeleteGameGenre(gomock.Any(), values.GameGenreIDFromUUID(testCase.genreID)).
				Return(testCase.DeleteGameGenreErr)

			c, _, rec := setupTestRequest(t, http.MethodDelete, fmt.Sprintf("/api/v2/genres/%s", testCase.genreID), nil)

			err := gameGenre.DeleteGameGenre(c, testCase.genreID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)
					} else {
						t.Errorf("error is not *echoHTTPError: %v", err)
					}
				} else if testCase.expectedErr != nil {
					assert.ErrorIs(t, err, testCase.expectedErr)
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestGetGameGenres(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreService := mock.NewMockGameGenre(ctrl)

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := session.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	gameGenre := NewGameGenre(mockGameGenreService, mock.NewMockGameV2(ctrl), session)

	type test struct {
		sessionExist     bool
		authSession      *domain.OIDCSession
		gameGenreInfos   []*service.GameGenreInfo
		GetGameGenresErr error
		apiGameGenres    []openapi.GameGenre
		isErr            bool
		statusCode       int
		expectedErr      error
	}

	gameGenreUUID1 := uuid.New()
	gameGenreID1 := values.GameGenreIDFromUUID(gameGenreUUID1)
	gameGenreNameStr1 := "ジャンル1"
	gameGenreName1 := values.NewGameGenreName(gameGenreNameStr1)

	gameGenreUUID2 := uuid.New()
	gameGenreID2 := values.GameGenreIDFromUUID(gameGenreUUID2)
	gameGenreNameStr2 := "ジャンル2"
	gameGenreName2 := values.NewGameGenreName(gameGenreNameStr2)

	now := time.Now()

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameGenreInfos: []*service.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, gameGenreName1, now.Add(-time.Hour)), Num: 2},
			},
			apiGameGenres: []openapi.GameGenre{
				{Id: gameGenreUUID1, Genre: gameGenreNameStr1, CreatedAt: now.Add(-time.Hour), Num: 2},
			},
			statusCode: http.StatusOK,
		},
		"sessionがあってもエラー無し": {
			sessionExist: true,
			authSession:  domain.NewOIDCSession(values.NewOIDCAccessToken("token"), now.Add(time.Hour)),
			gameGenreInfos: []*service.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, gameGenreName1, now.Add(-time.Hour)), Num: 2},
			},
			apiGameGenres: []openapi.GameGenre{
				{Id: gameGenreUUID1, Genre: gameGenreNameStr1, CreatedAt: now.Add(-time.Hour), Num: 2},
			},
			statusCode: http.StatusOK,
		},
		"複数あってもエラー無し": {
			gameGenreInfos: []*service.GameGenreInfo{
				{GameGenre: *domain.NewGameGenre(gameGenreID1, gameGenreName1, now.Add(-time.Hour)), Num: 2},
				{GameGenre: *domain.NewGameGenre(gameGenreID2, gameGenreName2, now.Add(-time.Hour*2)), Num: 3},
			},
			apiGameGenres: []openapi.GameGenre{
				{Id: gameGenreUUID1, Genre: gameGenreNameStr1, CreatedAt: now.Add(-time.Hour), Num: 2},
				{Id: gameGenreUUID2, Genre: gameGenreNameStr2, CreatedAt: now.Add(-time.Hour * 2), Num: 3},
			},
			statusCode: http.StatusOK,
		},
		"GetGameGenresがエラーなので500": {
			GetGameGenresErr: errors.New("test error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			mockGameGenreService.
				EXPECT().
				GetGameGenres(gomock.Any(), testCase.sessionExist).
				Return(testCase.gameGenreInfos, testCase.GetGameGenresErr)

			c, req, rec := setupTestRequest(t, http.MethodGet, "/api/v2/genres", nil)

			if testCase.sessionExist {
				setTestSession(t, c, req, rec, session, testCase.authSession)
			}

			err := gameGenre.GetGameGenres(c)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("err must be http error, but not http error: %v", err)
					}
				} else {
					if testCase.expectedErr != nil {
						assert.ErrorIs(t, err, testCase.expectedErr)
					} else {
						assert.Error(t, err)
					}
				}
			}

			if testCase.isErr || err != nil {
				return
			}

			assert.Equal(t, testCase.statusCode, rec.Code)

			var res []openapi.GameGenre
			err = json.NewDecoder(rec.Body).Decode(&res)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			assert.Len(t, res, len(testCase.apiGameGenres))
			for i := range res {
				assert.Equal(t, testCase.apiGameGenres[i].Id, res[i].Id)
				assert.Equal(t, testCase.apiGameGenres[i].Genre, res[i].Genre)
				assert.Equal(t, testCase.apiGameGenres[i].Num, res[i].Num)
				assert.WithinDuration(t, testCase.apiGameGenres[i].CreatedAt, res[i].CreatedAt, time.Second)
			}
		})
	}
}

func TestPutGameGenres(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreService := mock.NewMockGameGenre(ctrl)
	mockGameService := mock.NewMockGameV2(ctrl)

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := session.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	gameGenreHandler := NewGameGenre(mockGameGenreService, mockGameService, session)

	type test struct {
		gameID              openapi.GameGenreIDInPath
		invalidRequestBody  bool
		newGameGenreNames   openapi.PutGameGenresJSONRequestBody
		sessionExist        bool
		authSession         *domain.OIDCSession
		executeUpdateGame   bool
		UpdateGameGenresErr error
		executeGetGame      bool
		GetGameResult       *service.GameInfoV2
		GetGameErr          error
		isErr               bool
		statusCode          int
		res                 openapi.Game
		expectedErr         error
	}

	gameUUID := uuid.New()
	gameID := values.NewGameIDFromUUID(gameUUID)

	game := domain.NewGame(gameID, "name", "description", values.GameVisibilityTypePublic, time.Now())

	gameGenreNameStr1 := "ジャンル1"
	gameGenreNameStr2 := "ジャンル2"
	gameGenre1 := domain.NewGameGenre(values.NewGameGenreID(), values.NewGameGenreName(gameGenreNameStr1), time.Now())
	gameGenre2 := domain.NewGameGenre(values.NewGameGenreID(), values.NewGameGenreName(gameGenreNameStr2), time.Now())

	userNameStr1 := "ikura-hamu"
	userNameStr2 := "mazrean"
	user1 := service.NewUserInfo(values.NewTrapMemberID(uuid.New()), values.NewTrapMemberName(userNameStr1), values.TrapMemberStatusActive, false)
	user2 := service.NewUserInfo(values.NewTrapMemberID(uuid.New()), values.NewTrapMemberName(userNameStr2), values.TrapMemberStatusActive, false)

	testCases := map[string]test{
		"特に問題ないのでエラー無し": {
			gameID: gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{
				Genres: &[]string{gameGenreNameStr1, gameGenreNameStr2},
			},
			sessionExist:      true,
			authSession:       domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			executeUpdateGame: true,
			executeGetGame:    true,
			GetGameResult: &service.GameInfoV2{
				Game:        game,
				Genres:      []*domain.GameGenre{gameGenre1, gameGenre2},
				Owners:      []*service.UserInfo{user1},
				Maintainers: []*service.UserInfo{user2},
			},
			res: openapi.Game{
				Id:          gameUUID,
				Name:        "name",
				Description: "description",
				Visibility:  "public",
				CreatedAt:   time.Now(),
				Genres:      &[]string{gameGenreNameStr1, gameGenreNameStr2},
				Owners:      []string{userNameStr1},
				Maintainers: &[]string{userNameStr2},
			},
		},
		"リクエストボディがおかしいので400": {
			gameID:             gameUUID,
			invalidRequestBody: true,
			sessionExist:       true,
			authSession:        domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		"sessionが取得できないのでエラー": {
			gameID: gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{
				Genres: &[]string{gameGenreNameStr1, gameGenreNameStr2},
			},
			isErr:      true,
			statusCode: http.StatusUnauthorized,
		},
		"ジャンルが空でもエラー無し": {
			gameID:            gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{},
			sessionExist:      true,
			authSession:       domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			executeUpdateGame: true,
			executeGetGame:    true,
			GetGameResult: &service.GameInfoV2{
				Game:        game,
				Genres:      []*domain.GameGenre{},
				Owners:      []*service.UserInfo{user1},
				Maintainers: []*service.UserInfo{user2},
			},
			res: openapi.Game{
				Id:          gameUUID,
				Name:        "name",
				Description: "description",
				Visibility:  "public",
				CreatedAt:   time.Now(),
				Owners:      []string{userNameStr1},
				Maintainers: &[]string{userNameStr2},
			},
		},
		"ジャンルが長すぎるので400": {
			gameID: gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{
				Genres: &[]string{strings.Repeat("a", 100)},
			},
			sessionExist: true,
			authSession:  domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			isErr:        true,
			statusCode:   http.StatusBadRequest,
		},
		"UpdateGameGenresがErrNoGameなので404": {
			gameID: gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{
				Genres: &[]string{gameGenreNameStr1, gameGenreNameStr2},
			},
			sessionExist:        true,
			authSession:         domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			executeUpdateGame:   true,
			UpdateGameGenresErr: service.ErrNoGame,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		"UpdateGameGenresがErrDuplicateGameGenreなので400": {
			gameID: gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{
				Genres: &[]string{gameGenreNameStr1, gameGenreNameStr1},
			},
			sessionExist:        true,
			authSession:         domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			executeUpdateGame:   true,
			UpdateGameGenresErr: service.ErrDuplicateGameGenre,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		"UpdateGameGenresがエラーなので500": {
			gameID: gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{
				Genres: &[]string{gameGenreNameStr1, gameGenreNameStr1},
			},
			sessionExist:        true,
			authSession:         domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			executeUpdateGame:   true,
			UpdateGameGenresErr: errors.New("test error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
		"GetGameがエラーなので500": {
			gameID: gameUUID,
			newGameGenreNames: openapi.PutGameGenresJSONRequestBody{
				Genres: &[]string{gameGenreNameStr1, gameGenreNameStr1},
			},
			sessionExist:      true,
			authSession:       domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour)),
			executeUpdateGame: true,
			executeGetGame:    true,
			GetGameErr:        errors.New("test error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			if testCase.executeUpdateGame {
				mockGameGenreService.
					EXPECT().
					UpdateGameGenres(gomock.Any(), values.NewGameIDFromUUID(testCase.gameID), gomock.Any()).
					Return(testCase.UpdateGameGenresErr)
			}

			if testCase.executeGetGame {
				mockGameService.
					EXPECT().
					GetGame(gomock.Any(), gomock.Any(), values.NewGameIDFromUUID(testCase.gameID)).
					Return(testCase.GetGameResult, testCase.GetGameErr)
			}

			var bodyOpt bodyOpt
			if !testCase.invalidRequestBody {
				bodyOpt = withJSONBody(t, testCase.newGameGenreNames)
			} else {
				bodyOpt = withStringBody(t, "invalid request body")
			}

			c, req, rec := setupTestRequest(t, http.MethodPut, fmt.Sprintf("/api/v2/game/%s/genres", testCase.gameID), bodyOpt)

			if testCase.sessionExist {
				setTestSession(t, c, req, rec, session, testCase.authSession)
			}

			err = gameGenreHandler.PutGameGenres(c, testCase.gameID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("err must be http error, but not http error: %v", err)
					}
				} else {
					if testCase.expectedErr != nil {
						assert.ErrorIs(t, err, testCase.expectedErr)
					} else {
						assert.Error(t, err)
					}
				}
			}

			if testCase.isErr || err != nil {
				return
			}

			var res openapi.Game
			err = json.NewDecoder(rec.Body).Decode(&res)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			assert.Equal(t, testCase.res.Id, res.Id)
			assert.Equal(t, testCase.res.Name, res.Name)
			assert.Equal(t, testCase.res.Description, res.Description)
			assert.Equal(t, testCase.res.Visibility, res.Visibility)
			assert.WithinDuration(t, testCase.res.CreatedAt, res.CreatedAt, time.Second)
			if testCase.res.Genres != nil {
				assert.Len(t, *res.Genres, len(*testCase.res.Genres))
				for i := range *res.Genres {
					assert.Equal(t, (*testCase.res.Genres)[i], (*res.Genres)[i])
				}
			} else {
				assert.Nil(t, res.Genres)
			}
			assert.Len(t, res.Owners, len(testCase.res.Owners))
			for i := range res.Owners {
				assert.Equal(t, testCase.res.Owners[i], res.Owners[i])
			}
			if testCase.res.Maintainers != nil {
				assert.Len(t, *res.Maintainers, len(*testCase.res.Maintainers))
				for i := range *res.Maintainers {
					assert.Equal(t, (*testCase.res.Maintainers)[i], (*res.Maintainers)[i])
				}
			} else {
				assert.Nil(t, res.Maintainers)
			}
		})
	}
}

func TestPatchGameGenre(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameGenreService := mock.NewMockGameGenre(ctrl)
	mockGameService := mock.NewMockGameV2(ctrl)

	mockConf := mockConfig.NewMockHandler(ctrl)
	mockConf.
		EXPECT().
		SessionKey().
		Return("key", nil)
	mockConf.
		EXPECT().
		SessionSecret().
		Return("secret", nil)
	sess, err := session.NewSession(mockConf)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}
	session, err := NewSession(sess)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
		return
	}

	gameGenreHandler := NewGameGenre(mockGameGenreService, mockGameService, session)

	gameGenreID := uuid.New()

	testCases := map[string]struct {
		gameGenreID        openapi.GameGenreIDInPath
		req                openapi.PatchGameGenreJSONRequestBody
		invalidRequestBody bool
		executeUpdateGame  bool
		UpdateGameGenreErr error
		gameGenre          *service.GameGenreInfo
		resBody            openapi.GameGenre
		statusCode         int
		isErr              bool
		expectedErr        error
	}{
		"特に問題ないのでエラー無し": {
			gameGenreID: gameGenreID,
			req: openapi.PatchGameGenreJSONRequestBody{
				Genre: "new genre",
			},
			executeUpdateGame: true,
			gameGenre: &service.GameGenreInfo{
				GameGenre: *domain.NewGameGenre(values.GameGenreID(gameGenreID), values.NewGameGenreName("new genre"), time.Now()),
				Num:       1,
			},
			resBody: openapi.GameGenre{
				Id:        uuid.UUID(gameGenreID),
				Genre:     "new genre",
				CreatedAt: time.Now(),
				Num:       1,
			},
			statusCode: http.StatusOK,
		},
		"リクエストボディがおかしいので400": {
			gameGenreID:        gameGenreID,
			invalidRequestBody: true,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		"ジャンル名が空なので400": {
			gameGenreID: gameGenreID,
			req: openapi.PatchGameGenreJSONRequestBody{
				Genre: "",
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		"ジャンル名が長すぎるので400": {
			gameGenreID: gameGenreID,
			req: openapi.PatchGameGenreJSONRequestBody{
				Genre: strings.Repeat("a", 100),
			},
			isErr:      true,
			statusCode: http.StatusBadRequest,
		},
		"UpdateGameGenreがErrNoGameGenreなので404": {
			gameGenreID:        gameGenreID,
			req:                openapi.PatchGameGenreJSONRequestBody{Genre: "new genre"},
			executeUpdateGame:  true,
			UpdateGameGenreErr: service.ErrNoGameGenre,
			isErr:              true,
			statusCode:         http.StatusNotFound,
		},
		"UpdateGameGenreがErrDuplicateGameGenreNameなので400": {
			gameGenreID:        gameGenreID,
			req:                openapi.PatchGameGenreJSONRequestBody{Genre: "new genre"},
			executeUpdateGame:  true,
			UpdateGameGenreErr: service.ErrDuplicateGameGenreName,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		"UpdateGameGenreがErrNoGameGenreUpdatedなので400": {
			gameGenreID:        gameGenreID,
			req:                openapi.PatchGameGenreJSONRequestBody{Genre: "new genre"},
			executeUpdateGame:  true,
			UpdateGameGenreErr: service.ErrNoGameGenreUpdated,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		"UpdateGameGenreがエラーなので500": {
			gameGenreID:        gameGenreID,
			req:                openapi.PatchGameGenreJSONRequestBody{Genre: "new genre"},
			executeUpdateGame:  true,
			UpdateGameGenreErr: errors.New("test error"),
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			var bodyOpt bodyOpt
			if !testCase.invalidRequestBody {
				bodyOpt = withJSONBody(t, testCase.req)
			} else {
				bodyOpt = withStringBody(t, "invalid request body")
			}

			c, req, rec := setupTestRequest(t, http.MethodPatch, fmt.Sprintf("/api/v2/genres/%s", testCase.gameGenreID), bodyOpt)

			authSession := domain.NewOIDCSession(values.NewOIDCAccessToken("token"), time.Now().Add(time.Hour))
			setTestSession(t, c, req, rec, session, authSession)

			if testCase.executeUpdateGame {
				mockGameGenreService.
					EXPECT().
					UpdateGameGenre(gomock.Any(), values.GameGenreIDFromUUID(testCase.gameGenreID), values.NewGameGenreName(testCase.req.Genre)).
					Return(testCase.gameGenre, testCase.UpdateGameGenreErr)
			}

			err = gameGenreHandler.PatchGameGenre(c, testCase.gameGenreID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpErr *echo.HTTPError
					if errors.As(err, &httpErr) {
						assert.Equal(t, testCase.statusCode, httpErr.Code)
					} else {
						t.Errorf("err must be http error, but not http error: %v", err)
					}
				} else {
					if testCase.expectedErr != nil {
						assert.ErrorIs(t, err, testCase.expectedErr)
					} else {
						assert.Error(t, err)
					}
				}
			}

			if testCase.isErr {
				return
			}

			var res openapi.GameGenre
			err = json.NewDecoder(rec.Body).Decode(&res)
			require.NoError(t, err)

			assert.Equal(t, testCase.resBody.Id, res.Id)
			assert.Equal(t, testCase.resBody.Genre, res.Genre)
			assert.Equal(t, testCase.resBody.Num, res.Num)
			assert.WithinDuration(t, testCase.resBody.CreatedAt, res.CreatedAt, time.Second)
		})
	}
}
