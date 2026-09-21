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
	createdAt := time.Now().Truncate(time.Second)
	comment := values.NewFeedbackComment("comment")
	feedback := domain.NewGameFeedback(feedbackID, versionID, &comment, createdAt)
	answer := domain.NewGameFeedbackAnswer(values.NewGameFeedbackAnswerID(), feedbackID, questionID, 1)
	feedbackDetails := []*service.GameFeedbackDetail{
		{
			Feedback: feedback,
			Answers: []*service.GameFeedbackAnswerDetail{
				{
					Answer:       answer,
					QuestionText: values.NewFeedbackQuestionText("面白かったですか"),
					AnswerType:   values.FeedbackAnswerTypeYesNo,
				},
			},
		},
	}
	expectedResponse := openapi.GameFeedbacksResponse{
		Feedbacks: []openapi.GameFeedbackDetail{
			{
				Id:            openapi.GameFeedbackID(feedbackID),
				GameVersionID: openapi.GameVersionID(versionID),
				Answers: []openapi.FeedbackAnswer{
					feedbackAnswerYesNo(t, questionID, "面白かったですか", 1),
				},
				Comment:   ptrString("comment"),
				CreatedAt: createdAt,
			},
		},
		Total: 1,
	}

	testCases := map[string]struct {
		params                  openapi.GetGameFeedbacksParams
		executeGetGameFeedbacks bool
		expectedLimit           int
		expectedOffset          int
		getGameFeedbacksResult  []*service.GameFeedbackDetail
		getGameFeedbacksTotal   int
		getGameFeedbacksErr     error
		expectedResponse        *openapi.GameFeedbacksResponse
		isError                 bool
		statusCode              int
	}{
		"デフォルトのページングで正常に取得できる": {
			executeGetGameFeedbacks: true,
			expectedLimit:           50,
			expectedOffset:          0,
			getGameFeedbacksResult:  feedbackDetails,
			getGameFeedbacksTotal:   1,
			expectedResponse:        &expectedResponse,
			statusCode:              http.StatusOK,
		},
		"limitとoffsetを指定して正常に取得できる": {
			params: openapi.GetGameFeedbacksParams{
				Limit:  ptr(10),
				Offset: ptr(20),
			},
			executeGetGameFeedbacks: true,
			expectedLimit:           10,
			expectedOffset:          20,
			getGameFeedbacksResult:  feedbackDetails,
			getGameFeedbacksTotal:   1,
			expectedResponse:        &expectedResponse,
			statusCode:              http.StatusOK,
		},
		"limitが不正なので400": {
			params: openapi.GetGameFeedbacksParams{
				Limit: ptr(101),
			},
			isError:    true,
			statusCode: http.StatusBadRequest,
		},
		"offsetが不正なので400": {
			params: openapi.GetGameFeedbacksParams{
				Offset: ptr(-1),
			},
			isError:    true,
			statusCode: http.StatusBadRequest,
		},
		"GetGameFeedbacksがErrInvalidGameなので404": {
			executeGetGameFeedbacks: true,
			expectedLimit:           50,
			expectedOffset:          0,
			getGameFeedbacksErr:     service.ErrInvalidGame,
			isError:                 true,
			statusCode:              http.StatusNotFound,
		},
		"GetGameFeedbacksがErrInvalidLimitなので400": {
			executeGetGameFeedbacks: true,
			expectedLimit:           50,
			expectedOffset:          0,
			getGameFeedbacksErr:     service.ErrInvalidLimit,
			isError:                 true,
			statusCode:              http.StatusBadRequest,
		},
		"GetGameFeedbacksがエラーなので500": {
			executeGetGameFeedbacks: true,
			expectedLimit:           50,
			expectedOffset:          0,
			getGameFeedbacksErr:     assert.AnError,
			isError:                 true,
			statusCode:              http.StatusInternalServerError,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameFeedbackService := mock.NewMockGameFeedback(ctrl)
			handler := NewGameFeedback(gameFeedbackService)

			if testCase.executeGetGameFeedbacks {
				gameFeedbackService.
					EXPECT().
					GetGameFeedbacks(
						gomock.Any(),
						gameID,
						testCase.expectedLimit,
						testCase.expectedOffset,
					).
					Return(
						testCase.getGameFeedbacksResult,
						testCase.getGameFeedbacksTotal,
						testCase.getGameFeedbacksErr,
					)
			}

			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/games/%s/feedbacks", uuid.UUID(gameID)), nil)
			err := handler.GetGameFeedbacks(c, openapi.GameIDInPath(gameID), testCase.params)
			if testCase.isError {
				var httpError *echo.HTTPError
				require.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.statusCode, httpError.Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			var response openapi.GameFeedbacksResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			assert.Equal(t, *testCase.expectedResponse, response)
		})
	}
}

func ptr(value int) *int {
	return &value
}

func ptrString(value string) *string {
	return &value
}

func feedbackAnswerYesNo(t *testing.T, questionID values.FeedbackQuestionID, questionText string, answerValue int) openapi.FeedbackAnswer {
	t.Helper()

	answer := openapi.FeedbackAnswer{}
	require.NoError(t, answer.FromFeedbackAnswerYesNo(openapi.FeedbackAnswerYesNo{
		QuestionID:   openapi.FeedbackQuestionID(questionID),
		QuestionText: questionText,
		Answer:       answerValue,
	}))

	return answer
}
