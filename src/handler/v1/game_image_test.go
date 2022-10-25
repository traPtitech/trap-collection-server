package v1

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mockConf "github.com/traPtitech/trap-collection-server/src/config/mock"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestPostImage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mockConf.NewMockApp(ctrl)
	mockGameImageService := mock.NewMockGameImage(ctrl)

	gameImageHandler := NewGameImage(mockAppConfig, mockGameImageService)

	type test struct {
		description          string
		strGameID            string
		reader               *bytes.Reader
		executeSaveGameImage bool
		gameID               values.GameID
		SaveGameImageErr     error
		isErr                bool
		err                  error
		statusCode           int
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	testCases := []test{
		{
			description:          "特に問題ないのでエラーなし",
			strGameID:            uuid.UUID(gameID1).String(),
			reader:               bytes.NewReader([]byte("a")),
			executeSaveGameImage: true,
			gameID:               gameID1,
		},
		{
			description: "gameIDが不正なので400",
			strGameID:   "invalid",
			reader:      bytes.NewReader([]byte("a")),
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:          "SaveGameImageがErrInvalidGameIDなので400",
			strGameID:            uuid.UUID(gameID2).String(),
			reader:               bytes.NewReader([]byte("a")),
			executeSaveGameImage: true,
			gameID:               gameID2,
			SaveGameImageErr:     service.ErrInvalidGameID,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description:          "SaveGameImageがErrInvalidFormatなので400",
			strGameID:            uuid.UUID(gameID3).String(),
			reader:               bytes.NewReader([]byte("a")),
			executeSaveGameImage: true,
			gameID:               gameID3,
			SaveGameImageErr:     service.ErrInvalidFormat,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description:          "SaveGameImageがエラーなので500",
			strGameID:            uuid.UUID(gameID4).String(),
			reader:               bytes.NewReader([]byte("a")),
			executeSaveGameImage: true,
			gameID:               gameID4,
			SaveGameImageErr:     errors.New("error"),
			isErr:                true,
			statusCode:           http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/game/%s/image", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			r := &NopCloser{testCase.reader}

			if testCase.executeSaveGameImage {
				mockGameImageService.
					EXPECT().
					SaveGameImage(gomock.Any(), r, testCase.gameID).
					Return(testCase.SaveGameImageErr)
			}

			err := gameImageHandler.PostImage(c, testCase.strGameID, r)

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

func TestGetImage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAppConfig := mockConf.NewMockApp(ctrl)
	mockGameImageService := mock.NewMockGameImage(ctrl)

	gameImageHandler := NewGameImage(mockAppConfig, mockGameImageService)

	type test struct {
		description         string
		strGameID           string
		executeGetGameImage bool
		gameID              values.GameID
		tmpURL              values.GameImageTmpURL
		GetGameImageErr     error
		isErr               bool
		err                 error
		statusCode          int
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	urlLink, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to encode image: %v", err)
	}

	testCases := []test{
		{
			description:         "特に問題ないので303",
			strGameID:           uuid.UUID(gameID1).String(),
			executeGetGameImage: true,
			gameID:              gameID1,
			tmpURL:              values.NewGameImageTmpURL(urlLink),
			isErr:               true,
			statusCode:          http.StatusSeeOther,
		},
		{
			description: "gameIDが不正なので400",
			strGameID:   "invalid",
			isErr:       true,
			statusCode:  http.StatusBadRequest,
		},
		{
			description:         "GetGameImageがErrNoGameImageなので404",
			strGameID:           uuid.UUID(gameID2).String(),
			executeGetGameImage: true,
			gameID:              gameID2,
			GetGameImageErr:     service.ErrNoGameImage,
			isErr:               true,
			statusCode:          http.StatusNotFound,
		},
		{
			description:         "GetGameImageがErrInvalidGameIDなので400",
			strGameID:           uuid.UUID(gameID3).String(),
			executeGetGameImage: true,
			gameID:              gameID3,
			GetGameImageErr:     service.ErrInvalidGameID,
			isErr:               true,
			statusCode:          http.StatusBadRequest,
		},
		{
			description:         "GetGameImageがエラーなので500",
			strGameID:           uuid.UUID(gameID4).String(),
			executeGetGameImage: true,
			gameID:              gameID4,
			GetGameImageErr:     errors.New("error"),
			isErr:               true,
			statusCode:          http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/game/%s/image", testCase.strGameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if testCase.executeGetGameImage {
				mockGameImageService.
					EXPECT().
					GetGameImage(gomock.Any(), testCase.gameID).
					Return(testCase.tmpURL, testCase.GetGameImageErr)
			}

			err := gameImageHandler.GetImage(c, testCase.strGameID)

			if testCase.isErr {
				if testCase.statusCode != 0 {
					var httpError *echo.HTTPError
					if errors.As(err, &httpError) {
						assert.Equal(t, testCase.statusCode, httpError.Code)

						if testCase.statusCode == http.StatusSeeOther {
							assert.Equal(t, (*url.URL)(testCase.tmpURL).String(), c.Response().Header().Get(echo.HeaderLocation))
						}
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
