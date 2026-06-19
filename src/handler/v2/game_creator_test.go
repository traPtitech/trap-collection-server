package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetGameCreatorJobs(t *testing.T) {
	now := time.Now()
	gameID := uuid.UUID(values.NewGameID())

	presentJobID := values.NewGameCreatorJobID()
	customJobID := values.NewGameCreatorJobID()

	testCases := []struct {
		description string

		gameID openapi.GameIDInPath

		presentJobs []*domain.GameCreatorJob
		customJobs  []*domain.GameCreatorCustomJob
		serviceErr  error
		wantStatus  int
		wantErr     bool
		wantBody    []openapi.GameCreatorJob
	}{
		{
			description: "presentJobとcustomJobを取得できる",
			gameID:      gameID,
			presentJobs: []*domain.GameCreatorJob{
				domain.NewGameCreatorJob(
					presentJobID,
					values.NewGameCreatorJobDisplayName("Programmer"),
					now,
				),
			},
			customJobs: []*domain.GameCreatorCustomJob{
				domain.NewGameCreatorCustomJob(
					customJobID,
					values.NewGameCreatorJobDisplayName("customJob"),
					values.NewGameID(),
					now,
				),
			},
			wantStatus: http.StatusOK,
			wantBody: []openapi.GameCreatorJob{
				{
					Id:          uuid.UUID(presentJobID),
					DisplayName: "Programmer",
					IsCustomJob: false,
				},
				{
					Id:          uuid.UUID(customJobID),
					DisplayName: "customJob",
					IsCustomJob: true,
				},
			},
		},
		{
			description: "Jobが空欄でも取得できる",
			gameID:      uuid.UUID(values.NewGameID()),
			presentJobs: []*domain.GameCreatorJob{},
			customJobs:  []*domain.GameCreatorCustomJob{},
			wantStatus:  http.StatusOK,
			wantBody:    []openapi.GameCreatorJob{},
		},
		{
			description: "gameIDが不正なので404を返す",
			gameID:      uuid.UUID(values.NewGameID()),
			serviceErr:  service.ErrInvalidGameID,
			wantStatus:  http.StatusNotFound,
			wantErr:     true,
		},
		{
			description: "serviceがその他エラーなら500",
			gameID:      uuid.UUID(values.NewGameID()),
			serviceErr:  errors.New("error"),
			wantStatus:  http.StatusInternalServerError,
			wantErr:     true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockGameCreatorService := mock.NewMockGameCreator(ctrl)
			gameCreator := NewGameCreator(mockGameCreatorService)

			c, _, rec := setupTestRequest(
				t,
				http.MethodGet,
				fmt.Sprintf("/api/v2/games/%s/creators/jobs", testCase.gameID),
				nil,
			)

			mockGameCreatorService.
				EXPECT().
				GetGameCreatorJobs(gomock.Any(), values.NewGameIDFromUUID(testCase.gameID)).
				Return(testCase.presentJobs, testCase.customJobs, testCase.serviceErr)

			err := gameCreator.GetGameCreatorJobs(c, testCase.gameID)

			if testCase.wantErr {
				var httpError *echo.HTTPError
				if assert.ErrorAs(t, err, &httpError) {
					assert.Equal(t, testCase.wantStatus, httpError.Code)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.wantStatus, rec.Code)

			var res []openapi.GameCreatorJob
			err = json.NewDecoder(rec.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, testCase.wantBody, res)
		})
	}
}
