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

	questionID := values.NewFeedbackQuestionID()
	gameID := values.NewGameID()
	question := domain.NewFeedbackQuestion(questionID, gameID, values.NewFeedbackQuestionText("面白かったですか？"), values.FeedbackAnswerTypeYesNo, 0, time.Now(), nil)
	testCases := map[string]struct {
		serviceErr error
		wantStatus int
	}{
		"returns questions": {wantStatus: http.StatusOK},
		"missing game":      {serviceErr: service.ErrInvalidGame, wantStatus: http.StatusNotFound},
		"internal error":    {serviceErr: errors.New("unexpected"), wantStatus: http.StatusInternalServerError},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			feedbackService := mock.NewMockGameFeedback(ctrl)
			feedbackService.EXPECT().GetFeedbackQuestions(gomock.Any(), gameID).Return([]*domain.FeedbackQuestion{question}, testCase.serviceErr)
			handler := NewGameFeedback(feedbackService)
			c, _, rec := setupTestRequest(t, http.MethodGet, "/games/"+uuid.UUID(gameID).String()+"/feedback-questions", nil)
			err := handler.GetFeedbackQuestions(c, openapi.GameIDInPath(gameID))
			if testCase.wantStatus != http.StatusOK {
				var httpError *echo.HTTPError
				require.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.wantStatus, httpError.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			var response openapi.FeedbackQuestionsResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			assert.Equal(t, []openapi.FeedbackQuestion{{Id: questionID.UUID(), QuestionText: "面白かったですか？", AnswerType: openapi.AnswerTypeYesNo, QuestionOrder: 0}}, response.Questions)
		})
	}
}

func TestPutFeedbackQuestions(t *testing.T) {
	t.Parallel()
	gameID := values.NewGameID()
	questionID := values.NewFeedbackQuestionID()
	newQuestion := domain.NewFeedbackQuestion(questionID, gameID, values.NewFeedbackQuestionText("新しい質問"), values.FeedbackAnswerTypeFiveScale, 0, time.Now(), nil)

	t.Run("converts request and response", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		feedbackService := mock.NewMockGameFeedback(ctrl)
		feedbackService.EXPECT().PutFeedbackQuestions(gomock.Any(), gameID, gomock.Any()).DoAndReturn(func(_ context.Context, _ values.GameID, inputs []service.FeedbackQuestionInput) ([]*domain.FeedbackQuestion, error) {
			require.Len(t, inputs, 1)
			assert.Nil(t, inputs[0].ID)
			assert.Equal(t, values.NewFeedbackQuestionText("新しい質問"), inputs[0].QuestionText)
			assert.Equal(t, values.FeedbackAnswerTypeFiveScale, inputs[0].AnswerType)
			return []*domain.FeedbackQuestion{newQuestion}, nil
		})
		handler := NewGameFeedback(feedbackService)
		request := openapi.PutFeedbackQuestionsRequest{Questions: []openapi.FeedbackQuestionInput{{QuestionText: "新しい質問", AnswerType: openapi.AnswerTypeFiveScale}}}
		c, _, rec := setupTestRequest(t, http.MethodPut, "/games/"+uuid.UUID(gameID).String()+"/feedback-questions", withJSONBody(t, request))
		require.NoError(t, handler.PutFeedbackQuestions(c, openapi.GameIDInPath(gameID)))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects invalid answer type before service", func(t *testing.T) {
		handler := NewGameFeedback(mock.NewMockGameFeedback(gomock.NewController(t)))
		request := openapi.PutFeedbackQuestionsRequest{Questions: []openapi.FeedbackQuestionInput{{QuestionText: "質問", AnswerType: "invalid"}}}
		c, _, _ := setupTestRequest(t, http.MethodPut, "/games/"+uuid.UUID(gameID).String()+"/feedback-questions", withJSONBody(t, request))
		err := handler.PutFeedbackQuestions(c, openapi.GameIDInPath(gameID))
		var httpError *echo.HTTPError
		require.ErrorAs(t, err, &httpError)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	})

	t.Run("rejects omitted questions before service", func(t *testing.T) {
		handler := NewGameFeedback(mock.NewMockGameFeedback(gomock.NewController(t)))
		c, _, _ := setupTestRequest(t, http.MethodPut, "/games/"+uuid.UUID(gameID).String()+"/feedback-questions", withJSONBody(t, map[string]any{}))
		err := handler.PutFeedbackQuestions(c, openapi.GameIDInPath(gameID))
		var httpError *echo.HTTPError
		require.ErrorAs(t, err, &httpError)
		assert.Equal(t, http.StatusBadRequest, httpError.Code)
	})
}
