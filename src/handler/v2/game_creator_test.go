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

func TestGetGameCreators(t *testing.T) {
	t.Parallel()

	gameID := uuid.New()
	gameCreator := domain.NewGameCreatorWithJobs(
		domain.NewGameCreator(
			values.NewGameCreatorID(),
			values.NewTrapMemberID(uuid.New()),
			values.NewGameIDFromUUID(gameID),
			values.NewTrapMemberName("ikura-hamu"),
			time.Now(),
		),
		[]*domain.GameCreatorJob{
			domain.NewGameCreatorJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("job"), time.Now()),
		},
		[]*domain.GameCreatorCustomJob{
			domain.NewGameCreatorCustomJob(values.NewGameCreatorJobID(), values.NewGameCreatorJobDisplayName("custom job"), values.NewGameIDFromUUID(gameID), time.Now()),
		},
	)

	testCases := map[string]struct {
		gameID             openapi.GameIDInPath
		gameCreators       []*domain.GameCreatorWithJobs
		GetGameCreatorsErr error
		wantStatus         int
		wantBody           []openapi.GameCreator
		isError            bool
	}{
		"特に問題ないのでエラー無し": {
			gameID:       uuid.New(),
			gameCreators: []*domain.GameCreatorWithJobs{gameCreator},
			wantStatus:   http.StatusOK,
			wantBody: []openapi.GameCreator{
				{
					Jobs: []openapi.GameCreatorJob{
						{
							Id:          openapi.GameCreatorJobID(gameCreator.GetJobs()[0].GetID()),
							DisplayName: openapi.GameCreatorJobDisplayName(gameCreator.GetJobs()[0].GetDisplayName()),
							IsCustomJob: false,
						},
						{
							Id:          openapi.GameCreatorJobID(gameCreator.GetCustomJobs()[0].GetID()),
							DisplayName: openapi.GameCreatorJobDisplayName(gameCreator.GetCustomJobs()[0].GetDisplayName()),
							IsCustomJob: true,
						},
					},
					Name: openapi.UserName(gameCreator.GetGameCreator().GetUserName()),
				},
			},
		},
		"gameIDが不正なので404を返す": {
			gameID:             uuid.UUID(values.NewGameID()),
			GetGameCreatorsErr: service.ErrInvalidGameID,
			wantStatus:         http.StatusNotFound,
			isError:            true,
		},
		"GetGameCreatorsがエラーなので500": {
			gameID:             uuid.UUID(values.NewGameID()),
			GetGameCreatorsErr: assert.AnError,
			wantStatus:         http.StatusInternalServerError,
			isError:            true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockGameCreatorService := mock.NewMockGameCreator(ctrl)
			gc := NewGameCreator(mockGameCreatorService)

			mockGameCreatorService.EXPECT().
				GetGameCreators(gomock.Any(), values.NewGameIDFromUUID(testCase.gameID)).
				Return(testCase.gameCreators, testCase.GetGameCreatorsErr)

			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/games/%s/creators", testCase.gameID), nil)

			err := gc.GetGameCreators(c, testCase.gameID)

			if testCase.isError {
				var httpError *echo.HTTPError
				if assert.ErrorAs(t, err, &httpError) {
					assert.Equal(t, testCase.wantStatus, httpError.Code)
				}
				return
			}

			var resBody []openapi.GameCreator
			err = json.NewDecoder(rec.Body).Decode(&resBody)
			assert.NoError(t, err)
			assert.Equal(t, testCase.wantStatus, rec.Code)

			assert.Len(t, resBody, len(testCase.wantBody))
			for i, wantCreator := range testCase.wantBody {
				resCreator := resBody[i]

				assert.Equal(t, wantCreator.Name, resCreator.Name)
				assert.Len(t, resCreator.Jobs, len(wantCreator.Jobs))
				for j, wantJob := range wantCreator.Jobs {
					resJob := resCreator.Jobs[j]
					assert.Equal(t, wantJob, resJob)
				}
			}
		})
	}
}
