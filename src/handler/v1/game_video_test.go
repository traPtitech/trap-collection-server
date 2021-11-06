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

func TestPostVideo(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVideoService := mock.NewMockGameVideo(ctrl)

	gameVideoHandler := NewGameVideo(mockGameVideoService)

	type test struct {
		description          string
		strGameID            string
		reader               *bytes.Reader
		executeSaveGameVideo bool
		gameID               values.GameID
		SaveGameVideoErr     error
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
			executeSaveGameVideo: true,
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
			description:          "SaveGameVideoがErrInvalidGameIDなので400",
			strGameID:            uuid.UUID(gameID2).String(),
			reader:               bytes.NewReader([]byte("a")),
			executeSaveGameVideo: true,
			gameID:               gameID2,
			SaveGameVideoErr:     service.ErrInvalidGameID,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description:          "SaveGameVideoがErrInvalidFormatなので400",
			strGameID:            uuid.UUID(gameID3).String(),
			reader:               bytes.NewReader([]byte("a")),
			executeSaveGameVideo: true,
			gameID:               gameID3,
			SaveGameVideoErr:     service.ErrInvalidFormat,
			isErr:                true,
			statusCode:           http.StatusBadRequest,
		},
		{
			description:          "SaveGameVideoがエラーなので500",
			strGameID:            uuid.UUID(gameID4).String(),
			reader:               bytes.NewReader([]byte("a")),
			executeSaveGameVideo: true,
			gameID:               gameID4,
			SaveGameVideoErr:     errors.New("error"),
			isErr:                true,
			statusCode:           http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			r := &NopCloser{testCase.reader}

			if testCase.executeSaveGameVideo {
				mockGameVideoService.
					EXPECT().
					SaveGameVideo(gomock.Any(), r, testCase.gameID).
					Return(testCase.SaveGameVideoErr)
			}

			err := gameVideoHandler.PostVideo(testCase.strGameID, r)

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
