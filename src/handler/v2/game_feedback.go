package v2

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameFeedback struct {
	gameFeedbackService service.GameFeedback
}

func NewGameFeedback(gameFeedbackService service.GameFeedback) *GameFeedback {
	return &GameFeedback{
		gameFeedbackService: gameFeedbackService,
	}
}

// フィードバック設定の取得
// (GET /games/{gameID}/feedback-config)
func (gf *GameFeedback) GetFeedbackConfig(c echo.Context, gameID openapi.GameIDInPath) error {
	enabled, err := gf.gameFeedbackService.GetFeedbackConfig(
		c.Request().Context(),
		values.NewGameIDFromUUID(gameID),
	)
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if err != nil {
		log.Printf("error: failed to get feedback config: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get feedback config")
	}

	return c.JSON(http.StatusOK, openapi.FeedbackConfig{
		Enabled: enabled,
	})
}

// フィードバック設定の更新
// (PATCH /games/{gameID}/feedback-config)
func (gf *GameFeedback) PatchFeedbackConfig(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// フィードバック質問一覧の取得
// (GET /games/{gameID}/feedback-questions)
func (gf *GameFeedback) GetFeedbackQuestions(c echo.Context, gameID openapi.GameIDInPath) error {
	ctx := c.Request().Context()
	localGameID := values.NewGameIDFromUUID(gameID)

	questions, err := gf.gameFeedbackService.GetFeedbackQuestions(ctx, localGameID)
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if err != nil {
		log.Printf("error: failed to get feedback questions: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get feedback questions")
	}

	return c.JSON(http.StatusOK, feedbackQuestionsResponse(questions))
}

// フィードバック質問の一括設定
// (PUT /games/{gameID}/feedback-questions)
func (gf *GameFeedback) PutFeedbackQuestions(c echo.Context, gameID openapi.GameIDInPath) error {
	var request openapi.PutFeedbackQuestionsRequest
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if request.Questions == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "questions is required")
	}
	var rawRequest struct {
		Questions []struct {
			ID json.RawMessage `json:"id"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(body, &rawRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	inputs := make([]service.FeedbackQuestionInput, 0, len(request.Questions))
	for index, question := range request.Questions {
		if string(rawRequest.Questions[index].ID) == "null" {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid question ID")
		}
		answerType, ok := feedbackAnswerTypeFromOpenAPI(question.AnswerType)
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid answer type")
		}
		var id *values.FeedbackQuestionID
		if question.Id != nil {
			questionID := values.NewFeedbackQuestionIDFromUUID(*question.Id)
			id = &questionID
		}
		inputs = append(inputs, service.FeedbackQuestionInput{
			ID:           id,
			QuestionText: values.NewFeedbackQuestionText(question.QuestionText),
			AnswerType:   answerType,
		})
	}

	questions, err := gf.gameFeedbackService.PutFeedbackQuestions(c.Request().Context(), values.NewGameIDFromUUID(gameID), inputs)
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if errors.Is(err, service.ErrInvalidFeedbackQuestion) || errors.Is(err, service.ErrDuplicateFeedbackQuestion) || errors.Is(err, service.ErrInvalidFeedbackAnswerType) || errors.Is(err, service.ErrFeedbackQuestionAnswerTypeChange) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid feedback questions")
	}
	if err != nil {
		log.Printf("error: failed to put feedback questions: %v\\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to put feedback questions")
	}

	return c.JSON(http.StatusOK, feedbackQuestionsResponse(questions))
}

func feedbackQuestionsResponse(questions []*domain.FeedbackQuestion) openapi.FeedbackQuestionsResponse {
	responseQuestions := make([]openapi.FeedbackQuestion, 0, len(questions))
	for _, question := range questions {
		responseQuestions = append(responseQuestions, openapi.FeedbackQuestion{
			Id:            question.GetID().UUID(),
			QuestionText:  string(question.GetQuestionText()),
			AnswerType:    feedbackAnswerTypeToOpenAPI(question.GetAnswerType()),
			QuestionOrder: int(question.GetQuestionOrder()),
		})
	}
	return openapi.FeedbackQuestionsResponse{Questions: responseQuestions}
}

func feedbackAnswerTypeFromOpenAPI(answerType openapi.AnswerType) (values.FeedbackAnswerType, bool) {
	switch answerType {
	case openapi.AnswerTypeYesNo:
		return values.FeedbackAnswerTypeYesNo, true
	case openapi.AnswerTypeFiveScale:
		return values.FeedbackAnswerTypeFiveScale, true
	default:
		return 0, false
	}
}

func feedbackAnswerTypeToOpenAPI(answerType values.FeedbackAnswerType) openapi.AnswerType {
	switch answerType {
	case values.FeedbackAnswerTypeYesNo:
		return openapi.AnswerTypeYesNo
	case values.FeedbackAnswerTypeFiveScale:
		return openapi.AnswerTypeFiveScale
	default:
		return ""
	}
}

// ゲームフィードバックの送信
// (POST /games/{gameID}/feedbacks)
func (gf *GameFeedback) PostGameFeedback(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームのフィードバック一覧取得
// (GET /games/{gameID}/feedbacks)
func (gf *GameFeedback) GetGameFeedbacks(c echo.Context, _ openapi.GameIDInPath, _ openapi.GetGameFeedbacksParams) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームバージョンのフィードバック一覧取得
// (GET /games/{gameID}/versions/{gameVersionID}/feedbacks)
func (gf *GameFeedback) GetGameVersionFeedbacks(c echo.Context, _ openapi.GameIDInPath, _ openapi.GameVersionIDInPath, _ openapi.GetGameVersionFeedbacksParams) error {
	return c.NoContent(http.StatusNotImplemented)
}
