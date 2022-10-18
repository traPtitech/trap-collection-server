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
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
)

func TestGetGameVideos(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGameVideoService := mock.NewMockGameVideoV2(ctrl)

	gameVideo := NewGameVideo(mockGameVideoService)

	type test struct {
		description      string
		gameID           openapi.GameIDInPath
		videos           []*domain.GameVideo
		getGameVideosErr error
		resVideos        []openapi.GameVideo
		isErr            bool
		err              error
		statusCode       int
	}

	gameVideoID1 := values.NewGameVideoID()
	gameVideoID2 := values.NewGameVideoID()
	gameVideoID3 := values.NewGameVideoID()

	now := time.Now()
	testCases := []test{
		{
			description: "特に問題ないのでエラーなし",
			gameID:      uuid.UUID(values.NewGameID()),
			videos: []*domain.GameVideo{
				domain.NewGameVideo(
					gameVideoID1,
					values.GameVideoTypeMp4,
					now,
				),
			},
			resVideos: []openapi.GameVideo{
				{
					Id:        uuid.UUID(gameVideoID1),
					Mime:      openapi.Videomp4,
					CreatedAt: now,
				},
			},
		},
		{
			description: "mp4でないので500",
			gameID:      uuid.UUID(values.NewGameID()),
			videos: []*domain.GameVideo{
				domain.NewGameVideo(
					values.NewGameVideoID(),
					values.GameVideoType(100),
					now,
				),
			},
			isErr:      true,
			statusCode: http.StatusInternalServerError,
		},
		{
			description:      "GetGameVideosがErrInvalidGameIDなので404",
			gameID:           uuid.UUID(values.NewGameID()),
			getGameVideosErr: service.ErrInvalidGameID,
			isErr:            true,
			statusCode:       http.StatusNotFound,
		},
		{
			description:      "GetGameVideosがエラーなので500",
			gameID:           uuid.UUID(values.NewGameID()),
			getGameVideosErr: errors.New("error"),
			isErr:            true,
			statusCode:       http.StatusInternalServerError,
		},
		{
			description: "動画がなくても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			videos:      []*domain.GameVideo{},
			resVideos:   []openapi.GameVideo{},
		},
		{
			description: "動画が複数あっても問題なし",
			gameID:      uuid.UUID(values.NewGameID()),
			videos: []*domain.GameVideo{
				domain.NewGameVideo(
					gameVideoID2,
					values.GameVideoTypeMp4,
					now,
				),
				domain.NewGameVideo(
					gameVideoID3,
					values.GameVideoTypeMp4,
					now.Add(-10*time.Hour),
				),
			},
			resVideos: []openapi.GameVideo{
				{
					Id:        uuid.UUID(gameVideoID2),
					Mime:      openapi.Videomp4,
					CreatedAt: now,
				},
				{
					Id:        uuid.UUID(gameVideoID3),
					Mime:      openapi.Videomp4,
					CreatedAt: now.Add(-10 * time.Hour),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v2/games/%s/videos", testCase.gameID), nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockGameVideoService.
				EXPECT().
				GetGameVideos(gomock.Any(), gomock.Any()).
				Return(testCase.videos, testCase.getGameVideosErr)

			err := gameVideo.GetGameVideos(c, testCase.gameID)

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
			if err != nil || testCase.isErr {
				return
			}

			assert.Equal(t, http.StatusOK, rec.Code)

			var resVideos []openapi.GameVideo
			err = json.NewDecoder(rec.Body).Decode(&resVideos)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			for i, resVideo := range resVideos {
				assert.Equal(t, testCase.resVideos[i].Id, resVideo.Id)
				assert.Equal(t, testCase.resVideos[i].Mime, resVideo.Mime)
				assert.WithinDuration(t, testCase.resVideos[i].CreatedAt, resVideo.CreatedAt, time.Second)
			}
		})
	}
}
