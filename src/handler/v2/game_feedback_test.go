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
	feedbackWithoutComment := domain.NewGameFeedback(feedbackID, versionID, nil, createdAt)
	feedbackDetailsWithoutComment := []*service.GameFeedbackDetail{
		{
			Feedback: feedbackWithoutComment,
			Answers:  feedbackDetails[0].Answers,
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
	expectedResponseWithoutComment := expectedResponse
	expectedResponseWithoutComment.Feedbacks = append(
		[]openapi.GameFeedbackDetail(nil),
		expectedResponse.Feedbacks...,
	)
	expectedResponseWithoutComment.Feedbacks[0].Comment = nil

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
		"コメントがなくても正常に取得できる": {
			executeGetGameFeedbacks: true,
			expectedLimit:           50,
			expectedOffset:          0,
			getGameFeedbacksResult:  feedbackDetailsWithoutComment,
			getGameFeedbacksTotal:   1,
			expectedResponse:        &expectedResponseWithoutComment,
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

func TestGetGameVersionFeedbacks(t *testing.T) {
	gameID, versionID := values.NewGameID(), values.NewGameVersionID()
	feedbackID := values.NewGameFeedbackID()
	yesNoQuestionID, fiveScaleQuestionID := values.NewFeedbackQuestionID(), values.NewFeedbackQuestionID()
	comment := values.NewFeedbackComment("comment")
	createdAt := time.Now().Round(0)
	details := []*service.GameFeedbackDetail{
		{
			Feedback: domain.NewGameFeedback(
				feedbackID,
				versionID,
				&comment,
				createdAt,
			),
			Answers: []*service.GameFeedbackAnswerDetail{
				{
					Answer: domain.NewGameFeedbackAnswer(
						values.NewGameFeedbackAnswerID(),
						feedbackID,
						yesNoQuestionID,
						1,
					),
					QuestionText: values.NewFeedbackQuestionText("yes no"),
					AnswerType:   values.FeedbackAnswerTypeYesNo,
				},
				{
					Answer: domain.NewGameFeedbackAnswer(
						values.NewGameFeedbackAnswerID(),
						feedbackID,
						fiveScaleQuestionID,
						5,
					),
					QuestionText: values.NewFeedbackQuestionText("scale"),
					AnswerType:   values.FeedbackAnswerTypeFiveScale,
				},
			},
		},
	}
	testCases := map[string]struct {
		params                         openapi.GetGameVersionFeedbacksParams
		executeGetGameVersionFeedbacks bool
		getGameVersionFeedbacksResult  []*service.GameFeedbackDetail
		getGameVersionFeedbacksTotal   int
		getGameVersionFeedbacksErr     error
		expectedStatus                 int
	}{
		"GetGameVersionFeedbacksが成功するので200が返される": {
			params: openapi.GetGameVersionFeedbacksParams{
				Limit:  ptr(2),
				Offset: ptr(3),
			},
			executeGetGameVersionFeedbacks: true,
			getGameVersionFeedbacksResult:  details,
			getGameVersionFeedbacksTotal:   4,
			expectedStatus:                 http.StatusOK,
		},
		"GetGameVersionFeedbacksがErrInvalidGameなので404": {
			params: openapi.GetGameVersionFeedbacksParams{
				Limit:  ptr(2),
				Offset: ptr(3),
			},
			executeGetGameVersionFeedbacks: true,
			getGameVersionFeedbacksErr:     service.ErrInvalidGame,
			expectedStatus:                 http.StatusNotFound,
		},
		"GetGameVersionFeedbacksがErrInvalidGameVersionなので404": {
			params: openapi.GetGameVersionFeedbacksParams{
				Limit:  ptr(2),
				Offset: ptr(3),
			},
			executeGetGameVersionFeedbacks: true,
			getGameVersionFeedbacksErr:     service.ErrInvalidGameVersion,
			expectedStatus:                 http.StatusNotFound,
		},
		"GetGameVersionFeedbacksがエラーなので500": {
			params: openapi.GetGameVersionFeedbacksParams{
				Limit:  ptr(2),
				Offset: ptr(3),
			},
			executeGetGameVersionFeedbacks: true,
			getGameVersionFeedbacksErr:     assert.AnError,
			expectedStatus:                 http.StatusInternalServerError,
		},
		"limitが不正なので400が返される": {
			params: openapi.GetGameVersionFeedbacksParams{
				Limit: ptr(101),
			},
			expectedStatus: http.StatusBadRequest,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gameFeedbackService := mock.NewMockGameFeedback(ctrl)
			handler := NewGameFeedback(gameFeedbackService)

			if testCase.executeGetGameVersionFeedbacks {
				gameFeedbackService.EXPECT().
					GetGameVersionFeedbacks(gomock.Any(), gameID, versionID, 2, 3).
					Return(
						testCase.getGameVersionFeedbacksResult,
						testCase.getGameVersionFeedbacksTotal,
						testCase.getGameVersionFeedbacksErr,
					)
			}

			c, _, rec := setupTestRequest(t, http.MethodGet, fmt.Sprintf("/games/%s/versions/%s/feedbacks", uuid.UUID(gameID), uuid.UUID(versionID)), nil)
			err := handler.GetGameVersionFeedbacks(c, openapi.GameIDInPath(gameID), openapi.GameVersionIDInPath(versionID), testCase.params)
			if testCase.expectedStatus != http.StatusOK {
				var httpError *echo.HTTPError
				require.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.expectedStatus, httpError.Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			var response openapi.GameVersionFeedbacksResponse
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
			assert.Equal(t, 4, response.Total)
			require.Len(t, response.Feedbacks, 1)
			assert.Equal(t, openapi.GameFeedbackID(feedbackID), response.Feedbacks[0].Id)
			assert.Equal(t, createdAt, response.Feedbacks[0].CreatedAt)
			require.NotNil(t, response.Feedbacks[0].Comment)
			assert.Equal(t, "comment", *response.Feedbacks[0].Comment)
			require.Len(t, response.Feedbacks[0].Answers, 2)
			yesNo, parseErr := response.Feedbacks[0].Answers[0].AsFeedbackAnswerYesNo()
			require.NoError(t, parseErr)
			assert.Equal(t, openapi.FeedbackQuestionID(yesNoQuestionID), yesNo.QuestionID)
			assert.Equal(t, 1, yesNo.Answer)
			fiveScale, parseErr := response.Feedbacks[0].Answers[1].AsFeedbackAnswerFiveScale()
			require.NoError(t, parseErr)
			assert.Equal(t, openapi.FeedbackQuestionID(fiveScaleQuestionID), fiveScale.QuestionID)
			assert.Equal(t, 5, fiveScale.Answer)
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
