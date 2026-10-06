package v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestGetFeedbackConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		enabled    bool
		serviceErr error
		wantStatus int
		wantErr    bool
	}{
		"有効な設定を取得できる": {
			enabled:    true,
			wantStatus: http.StatusOK,
		},
		"無効な設定を取得できる": {
			enabled:    false,
			wantStatus: http.StatusOK,
		},
		"ゲームが存在しないので404": {
			serviceErr: service.ErrInvalidGame,
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
		"serviceがその他のエラーなので500": {
			serviceErr: errors.New("unexpected error"),
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameFeedbackService := mock.NewMockGameFeedback(ctrl)
			handler := NewGameFeedback(gameFeedbackService)
			gameID := values.NewGameID()

			gameFeedbackService.
				EXPECT().
				GetFeedbackConfig(gomock.Any(), gameID).
				Return(testCase.enabled, testCase.serviceErr)

			c, _, rec := setupTestRequest(
				t,
				http.MethodGet,
				fmt.Sprintf("/games/%s/feedback-config", uuid.UUID(gameID).String()),
				nil,
			)

			err := handler.GetFeedbackConfig(c, openapi.GameIDInPath(gameID))
			if testCase.wantErr {
				var httpError *echo.HTTPError
				require.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.wantStatus, httpError.Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.wantStatus, rec.Code)

			var response openapi.FeedbackConfig
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			assert.Equal(t, openapi.FeedbackConfig{Enabled: testCase.enabled}, response)
		})
	}
}

func TestGetFeedbackQuestions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gameID := values.NewGameID()
	questionID := values.NewFeedbackQuestionID()
	question := domain.NewFeedbackQuestion(
		questionID,
		gameID,
		values.NewFeedbackQuestionText("面白かったですか？"),
		values.FeedbackAnswerTypeYesNo,
		values.NewFeedbackQuestionOrder(0),
		time.Now(),
		nil,
	)
	expectedQuestions := []*domain.FeedbackQuestion{
		question,
	}
	expectedResponse := openapi.FeedbackQuestionsResponse{
		Questions: []openapi.FeedbackQuestion{
			{
				Id:            questionID.UUID(),
				QuestionText:  "面白かったですか？",
				AnswerType:    openapi.AnswerTypeYesNo,
				QuestionOrder: 0,
			},
		},
	}

	testCases := map[string]struct {
		gameID                      values.GameID
		executeGetFeedbackQuestions bool
		getFeedbackQuestionsResult  []*domain.FeedbackQuestion
		getFeedbackQuestionsErr     error
		expectedResponse            openapi.FeedbackQuestionsResponse
		isError                     bool
		statusCode                  int
	}{
		"GetFeedbackQuestionsが成功するので200": {
			gameID:                      gameID,
			executeGetFeedbackQuestions: true,
			getFeedbackQuestionsResult:  expectedQuestions,
			expectedResponse:            expectedResponse,
			statusCode:                  http.StatusOK,
		},
		"GetFeedbackQuestionsがErrInvalidGameなので404": {
			gameID:                      gameID,
			executeGetFeedbackQuestions: true,
			getFeedbackQuestionsErr:     service.ErrInvalidGame,
			isError:                     true,
			statusCode:                  http.StatusNotFound,
		},
		"GetFeedbackQuestionsがエラーなので500": {
			gameID:                      gameID,
			executeGetFeedbackQuestions: true,
			getFeedbackQuestionsErr:     errors.New("unexpected"),
			isError:                     true,
			statusCode:                  http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameFeedbackService := mock.NewMockGameFeedback(ctrl)
			handler := NewGameFeedback(gameFeedbackService)

			if testCase.executeGetFeedbackQuestions {
				gameFeedbackService.
					EXPECT().
					GetFeedbackQuestions(ctx, testCase.gameID).
					Return(testCase.getFeedbackQuestionsResult, testCase.getFeedbackQuestionsErr)
			}

			url := fmt.Sprintf("/games/%s/feedback-questions", uuid.UUID(testCase.gameID))
			c, request, rec := setupTestRequest(t, http.MethodGet, url, nil)
			request = request.WithContext(ctx)
			c.SetRequest(request)

			err := handler.GetFeedbackQuestions(c, openapi.GameIDInPath(testCase.gameID))

			if testCase.isError {
				var httpError *echo.HTTPError
				if assert.ErrorAs(t, err, &httpError) {
					assert.Equal(t, testCase.statusCode, httpError.Code)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			var response openapi.FeedbackQuestionsResponse
			assert.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			assert.Equal(t, testCase.expectedResponse, response)
		})
	}
}
