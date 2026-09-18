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

func TestGetGameFeedbacks(t *testing.T) {
	t.Parallel()
	gameID := values.NewGameID()
	feedbackID := values.NewGameFeedbackID()
	versionID := values.NewGameVersionID()
	questionID := values.NewFeedbackQuestionID()
	feedback := domain.NewGameFeedback(feedbackID, versionID, nil, time.Now())
	answer := domain.NewGameFeedbackAnswer(values.NewGameFeedbackAnswerID(), feedbackID, questionID, 1)
	details := []*service.GameFeedbackDetail{{Feedback: feedback, Answers: []*service.GameFeedbackAnswerDetail{{Answer: answer, QuestionText: values.NewFeedbackQuestionText("面白かったですか"), AnswerType: values.FeedbackAnswerTypeYesNo}}}}
	testCases := map[string]struct {
		params     openapi.GetGameFeedbacksParams
		serviceErr error
		wantStatus int
		wantErr    bool
	}{
		"デフォルトのページングで取得できる": {wantStatus: http.StatusOK},
		"gameが存在しない":        {serviceErr: service.ErrInvalidGame, wantStatus: http.StatusNotFound, wantErr: true},
		"limitが不正":          {params: openapi.GetGameFeedbacksParams{Limit: ptr(101)}, wantStatus: http.StatusBadRequest, wantErr: true},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			serviceMock := mock.NewMockGameFeedback(ctrl)
			handler := NewGameFeedback(serviceMock)
			if testCase.params.Limit == nil {
				serviceMock.EXPECT().GetGameFeedbacks(gomock.Any(), gameID, 50, 0).Return(details, 1, testCase.serviceErr)
			}
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/games/%s/feedbacks", uuid.UUID(gameID)), nil)
			err := handler.GetGameFeedbacks(c, openapi.GameIDInPath(gameID), testCase.params)
			if testCase.wantErr {
				var httpError *echo.HTTPError
				require.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.wantStatus, httpError.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.wantStatus, rec.Code)
			var response openapi.GameFeedbacksResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			require.Len(t, response.Feedbacks, 1)
			assert.NotNil(t, response.Feedbacks[0].Answers)
		})
	}
}

func TestGetGameVersionFeedbacks(t *testing.T) {
	gameID, versionID := values.NewGameID(), values.NewGameVersionID()
	feedbackID, questionID := values.NewGameFeedbackID(), values.NewFeedbackQuestionID()
	details := []*service.GameFeedbackDetail{{Feedback: domain.NewGameFeedback(feedbackID, versionID, nil, time.Now()), Answers: []*service.GameFeedbackAnswerDetail{{Answer: domain.NewGameFeedbackAnswer(values.NewGameFeedbackAnswerID(), feedbackID, questionID, 5), QuestionText: values.NewFeedbackQuestionText("評価"), AnswerType: values.FeedbackAnswerTypeFiveScale}}}}
	testCases := map[string]struct {
		err    error
		status int
		call   bool
	}{
		"取得できる":         {status: http.StatusOK, call: true},
		"gameが存在しない":    {err: service.ErrInvalidGame, status: http.StatusNotFound, call: true},
		"versionが存在しない": {err: service.ErrInvalidGameVersion, status: http.StatusNotFound, call: true},
		"下位エラー":         {err: assert.AnError, status: http.StatusInternalServerError, call: true},
		"limitが不正":      {status: http.StatusBadRequest},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			serviceMock := mock.NewMockGameFeedback(gomock.NewController(t))
			handler := NewGameFeedback(serviceMock)
			params := openapi.GetGameVersionFeedbacksParams{Limit: ptr(2), Offset: ptr(3)}
			if testCase.call {
				serviceMock.EXPECT().GetGameVersionFeedbacks(gomock.Any(), gameID, versionID, 2, 3).Return(details, 1, testCase.err)
			} else {
				params.Limit = ptr(101)
			}
			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/games/%s/versions/%s/feedbacks", uuid.UUID(gameID), uuid.UUID(versionID)), nil)
			err := handler.GetGameVersionFeedbacks(c, openapi.GameIDInPath(gameID), openapi.GameVersionIDInPath(versionID), params)
			if testCase.status != http.StatusOK {
				var httpError *echo.HTTPError
				require.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.status, httpError.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func ptr(value int) *int {
	return &value
}
