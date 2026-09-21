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

func TestPutFeedbackQuestions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	gameID := values.NewGameID()
	questionID := values.NewFeedbackQuestionID()
	newQuestion := domain.NewFeedbackQuestion(
		questionID,
		gameID,
		values.NewFeedbackQuestionText("新しい質問"),
		values.FeedbackAnswerTypeFiveScale,
		values.NewFeedbackQuestionOrder(0),
		time.Now(),
		nil,
	)
	expectedResponse := openapi.FeedbackQuestionsResponse{
		Questions: []openapi.FeedbackQuestion{{
			Id:            questionID.UUID(),
			QuestionText:  "新しい質問",
			AnswerType:    openapi.AnswerTypeFiveScale,
			QuestionOrder: 0,
		}},
	}

	testCases := map[string]struct {
		requestBody any

		executePutFeedbackQuestions bool
		expectedInputs              []service.FeedbackQuestionInput
		putFeedbackQuestionsResult  []*domain.FeedbackQuestion
		putFeedbackQuestionsErr     error

		expectedResponse openapi.FeedbackQuestionsResponse
		isError          bool
		statusCode       int
	}{
		"正常にリクエストを変換してレスポンスを返す": {
			requestBody: openapi.PutFeedbackQuestionsRequest{
				Questions: []openapi.FeedbackQuestionInput{{
					QuestionText: "新しい質問",
					AnswerType:   openapi.AnswerTypeFiveScale,
				}},
			},
			executePutFeedbackQuestions: true,
			expectedInputs: []service.FeedbackQuestionInput{{
				QuestionText: values.NewFeedbackQuestionText("新しい質問"),
				AnswerType:   values.FeedbackAnswerTypeFiveScale,
			}},
			putFeedbackQuestionsResult: []*domain.FeedbackQuestion{newQuestion},
			expectedResponse:           expectedResponse,
			statusCode:                 http.StatusOK,
		},
		"questionsが空配列なので正常に空配列を返す": {
			requestBody: openapi.PutFeedbackQuestionsRequest{
				Questions: []openapi.FeedbackQuestionInput{},
			},
			executePutFeedbackQuestions: true,
			expectedInputs:              []service.FeedbackQuestionInput{},
			putFeedbackQuestionsResult:  []*domain.FeedbackQuestion{},
			expectedResponse: openapi.FeedbackQuestionsResponse{
				Questions: []openapi.FeedbackQuestion{},
			},
			statusCode: http.StatusOK,
		},
		"answerTypeが不正なので400": {
			requestBody: openapi.PutFeedbackQuestionsRequest{
				Questions: []openapi.FeedbackQuestionInput{{
					QuestionText: "質問",
					AnswerType:   "invalid",
				}},
			},
			isError:    true,
			statusCode: http.StatusBadRequest,
		},
		"questionsが未指定なので400": {
			requestBody: map[string]any{},
			isError:     true,
			statusCode:  http.StatusBadRequest,
		},
		"questionsがnullなので400": {
			requestBody: map[string]any{"questions": nil},
			isError:     true,
			statusCode:  http.StatusBadRequest,
		},
		"idがnullなので400": {
			requestBody: map[string]any{
				"questions": []map[string]any{{
					"id":           nil,
					"questionText": "質問",
					"answerType":   "yesNo",
				}},
			},
			isError:    true,
			statusCode: http.StatusBadRequest,
		},
	}

	for description, testCase := range testCases {
		t.Run(description, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			gameFeedbackService := mock.NewMockGameFeedback(ctrl)
			handler := NewGameFeedback(gameFeedbackService)

			if testCase.executePutFeedbackQuestions {
				gameFeedbackService.
					EXPECT().
					PutFeedbackQuestions(ctx, gameID, testCase.expectedInputs).
					Return(testCase.putFeedbackQuestionsResult, testCase.putFeedbackQuestionsErr)
			}

			url := fmt.Sprintf("/games/%s/feedback-questions", uuid.UUID(gameID))
			c, request, recorder := setupTestRequest(
				t,
				http.MethodPut,
				url,
				withJSONBody(t, testCase.requestBody),
			)
			request = request.WithContext(ctx)
			c.SetRequest(request)

			err := handler.PutFeedbackQuestions(c, openapi.GameIDInPath(gameID))
			if testCase.isError {
				var httpError *echo.HTTPError
				if assert.ErrorAs(t, err, &httpError) {
					assert.Equal(t, testCase.statusCode, httpError.Code)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, recorder.Code)

			var response openapi.FeedbackQuestionsResponse
			assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
			assert.Equal(t, testCase.expectedResponse, response)
		})
	}
}
