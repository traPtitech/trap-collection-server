package v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestPostURL(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameURLService := mock.NewMockGameURL(ctrl)

	gameURLHandler := NewGameURL(mockGameURLService)

	type test struct {
		description        string
		strGameID          string
		strGameURL         string
		executeSaveGameURL bool
		gameID             values.GameID
		gameURL            *domain.GameURL
		SaveGameURLErr     error
		apiGameURL         *openapi.GameUrl
		isErr              bool
		err                error
		statusCode         int
	}

	gameID := values.NewGameID()
	gameURLID := values.NewGameURLID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}
	link := values.NewGameURLLink(urlLink)

	testCases := []test{
		{
			description:        "特に問題ないのでエラーなし",
			strGameID:          uuid.UUID(gameID).String(),
			strGameURL:         "https://example.com",
			executeSaveGameURL: true,
			gameID:             gameID,
			gameURL: domain.NewGameURL(
				gameURLID,
				link,
				time.Now(),
			),
			apiGameURL: &openapi.GameUrl{
				Id:  uuid.UUID(gameURLID).String(),
				Url: (*url.URL)(link).String(),
			},
		},
		{
			description: "gameIDが不正なので400",
			strGameID:   "invalid",
			strGameURL:  "https://example.com",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description: "urlが不正なので400",
			strGameID:   uuid.UUID(gameID).String(),
			strGameURL:  " https://example.com",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:        "SaveGameURLがErrInvalidGameIDなので400",
			strGameID:          uuid.UUID(gameID).String(),
			strGameURL:         "https://example.com",
			executeSaveGameURL: true,
			gameID:             gameID,
			SaveGameURLErr:     service.ErrInvalidGameID,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "SaveGameURLがErrNoGameVersionなので400",
			strGameID:          uuid.UUID(gameID).String(),
			strGameURL:         "https://example.com",
			executeSaveGameURL: true,
			gameID:             gameID,
			SaveGameURLErr:     service.ErrNoGameVersion,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "SaveGameURLがErrGameURLAlreadyExistsなので400",
			strGameID:          uuid.UUID(gameID).String(),
			strGameURL:         "https://example.com",
			executeSaveGameURL: true,
			gameID:             gameID,
			SaveGameURLErr:     service.ErrGameURLAlreadyExists,
			isErr:              true,
			statusCode:         http.StatusBadRequest,
		},
		{
			description:        "SaveGameURLがエラーなので500",
			strGameID:          uuid.UUID(gameID).String(),
			strGameURL:         "https://example.com",
			executeSaveGameURL: true,
			gameID:             gameID,
			SaveGameURLErr:     errors.New("error"),
			isErr:              true,
			statusCode:         http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/game/%s/asset/url", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeSaveGameURL {
				mockGameURLService.
					EXPECT().
					SaveGameURL(gomock.Any(), testCase.gameID, gomock.Any()).
					Return(testCase.gameURL, testCase.SaveGameURLErr)
			}

			gameURL, err := gameURLHandler.PostURL(c, testCase.strGameID, &openapi.NewGameUrl{
				Url: testCase.strGameURL,
			})

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

			assert.Equal(t, *testCase.apiGameURL, *gameURL)
		})
	}
}

func TestGetGameURL(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameURLService := mock.NewMockGameURL(ctrl)

	gameURLHandler := NewGameURL(mockGameURLService)

	type test struct {
		description       string
		strGameID         string
		executeGetGameURL bool
		gameID            values.GameID
		gameURL           *domain.GameURL
		GetGameURLErr     error
		apiGameURL        string
		isErr             bool
		err               error
		statusCode        int
	}

	gameID := values.NewGameID()
	gameURLID := values.NewGameURLID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}
	link := values.NewGameURLLink(urlLink)

	testCases := []test{
		{
			description:       "特に問題ないのでエラーなし",
			strGameID:         uuid.UUID(gameID).String(),
			executeGetGameURL: true,
			gameID:            gameID,
			gameURL: domain.NewGameURL(
				gameURLID,
				link,
				time.Now(),
			),
			apiGameURL: (*url.URL)(link).String(),
		},
		{
			description: "gameIDが不正なので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:       "GetGameURLがErrInvalidGameIDなので400",
			strGameID:         uuid.UUID(gameID).String(),
			executeGetGameURL: true,
			gameID:            gameID,
			GetGameURLErr:     service.ErrInvalidGameID,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "GetGameURLがErrNoGameVersionなので400",
			strGameID:         uuid.UUID(gameID).String(),
			executeGetGameURL: true,
			gameID:            gameID,
			GetGameURLErr:     service.ErrNoGameVersion,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "GetGameURLがErrNoGameURLなので400",
			strGameID:         uuid.UUID(gameID).String(),
			executeGetGameURL: true,
			gameID:            gameID,
			GetGameURLErr:     service.ErrNoGameURL,
			isErr:             true,
			statusCode:        http.StatusBadRequest,
		},
		{
			description:       "GetGameURLがエラーなので500",
			strGameID:         uuid.UUID(gameID).String(),
			executeGetGameURL: true,
			gameID:            gameID,
			GetGameURLErr:     errors.New("error"),
			isErr:             true,
			statusCode:        http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/game/%s/asset/url", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeGetGameURL {
				mockGameURLService.
					EXPECT().
					GetGameURL(gomock.Any(), testCase.gameID).
					Return(testCase.gameURL, testCase.GetGameURLErr)
			}

			gameURL, err := gameURLHandler.GetGameURL(c, testCase.strGameID)

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

			assert.Equal(t, testCase.apiGameURL, gameURL)
		})
	}
}
