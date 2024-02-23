package v2

import (
	"encoding/json"
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
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
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

	gameGenre := NewGameGenre(mockGameGenreService, session)

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

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v2/genres/%s", testCase.genreID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

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

	gameGenre := NewGameGenre(mockGameGenreService, session)

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

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v2/genres", nil)
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
