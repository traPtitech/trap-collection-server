package v1

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestPostImage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameImageService := mock.NewMockGameImage(ctrl)

	gameImageHandler := NewGameImage(mockGameImageService)

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
			r := &NopCloser{testCase.reader}

			if testCase.executeSaveGameImage {
				mockGameImageService.
					EXPECT().
					SaveGameImage(gomock.Any(), r, testCase.gameID).
					Return(testCase.SaveGameImageErr)
			}

			err := gameImageHandler.PostImage(testCase.strGameID, r)

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

	mockGameImageService := mock.NewMockGameImage(ctrl)

	gameImageHandler := NewGameImage(mockGameImageService)

	type test struct {
		description         string
		strGameID           string
		executeGetGameImage bool
		gameID              values.GameID
		GetGameImageErr     error
		isErr               bool
		err                 error
		statusCode          int
	}

	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()
	gameID3 := values.NewGameID()
	gameID4 := values.NewGameID()

	testCases := []test{
		{
			description:         "特に問題ないのでエラーなし",
			strGameID:           uuid.UUID(gameID1).String(),
			executeGetGameImage: true,
			gameID:              gameID1,
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
			r := bytes.NewReader([]byte("a"))

			if testCase.executeGetGameImage {
				mockGameImageService.
					EXPECT().
					GetGameImage(gomock.Any(), testCase.gameID).
					Return(r, testCase.GetGameImageErr)
			}

			res, err := gameImageHandler.GetImage(testCase.strGameID)

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

			assert.Equal(t, r, res)
		})
	}
}
