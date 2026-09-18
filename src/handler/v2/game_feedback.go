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
	questions, err := gf.gameFeedbackService.GetFeedbackQuestions(c.Request().Context(), values.NewGameIDFromUUID(gameID))
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if err != nil {
		log.Printf("error: failed to get feedback questions: %v\\n", err)
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
func (gf *GameFeedback) GetGameFeedbacks(c echo.Context, gameIDPath openapi.GameIDInPath, params openapi.GetGameFeedbacksParams) error {
	limit, offset, err := gameFeedbackPagination(params.Limit, params.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pagination")
	}

	feedbacks, total, err := gf.gameFeedbackService.GetGameFeedbacks(
		c.Request().Context(),
		values.NewGameIDFromUUID(gameIDPath),
		limit,
		offset,
	)
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if errors.Is(err, service.ErrInvalidLimit) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pagination")
	}
	if err != nil {
		log.Printf("error: failed to get game feedbacks: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game feedbacks")
	}

	response := openapi.GameFeedbacksResponse{
		Feedbacks: make([]openapi.GameFeedbackDetail, 0, len(feedbacks)),
		Total:     total,
	}
	for _, feedback := range feedbacks {
		answers, err := feedbackAnswersToOpenAPI(feedback.Answers)
		if err != nil {
			log.Printf("error: failed to convert feedback answers: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game feedbacks")
		}

		var comment *string
		if feedback.Feedback.GetComment() != nil {
			value := string(*feedback.Feedback.GetComment())
			comment = &value
		}
		response.Feedbacks = append(response.Feedbacks, openapi.GameFeedbackDetail{
			Id:            openapi.GameFeedbackID(feedback.Feedback.GetID()),
			GameVersionID: openapi.GameVersionID(feedback.Feedback.GetGameVersionID()),
			Answers:       answers,
			Comment:       comment,
			CreatedAt:     feedback.Feedback.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// ゲームバージョンのフィードバック一覧取得
// (GET /games/{gameID}/versions/{gameVersionID}/feedbacks)
func (gf *GameFeedback) GetGameVersionFeedbacks(c echo.Context, gameIDPath openapi.GameIDInPath, gameVersionIDPath openapi.GameVersionIDInPath, params openapi.GetGameVersionFeedbacksParams) error {
	limit, offset, err := gameFeedbackPagination(params.Limit, params.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pagination")
	}

	feedbacks, total, err := gf.gameFeedbackService.GetGameVersionFeedbacks(
		c.Request().Context(),
		values.NewGameIDFromUUID(gameIDPath),
		values.NewGameVersionIDFromUUID(gameVersionIDPath),
		limit,
		offset,
	)
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if errors.Is(err, service.ErrInvalidGameVersion) {
		return echo.NewHTTPError(http.StatusNotFound, "game version not found")
	}
	if errors.Is(err, service.ErrInvalidLimit) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid pagination")
	}
	if err != nil {
		log.Printf("error: failed to get game version feedbacks: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game version feedbacks")
	}

	response := openapi.GameVersionFeedbacksResponse{
		Feedbacks: make([]openapi.FeedbackDetail, 0, len(feedbacks)),
		Total:     total,
	}
	for _, feedback := range feedbacks {
		answers, err := feedbackAnswersToOpenAPI(feedback.Answers)
		if err != nil {
			log.Printf("error: failed to convert feedback answers: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game version feedbacks")
		}

		var comment *string
		if feedback.Feedback.GetComment() != nil {
			value := string(*feedback.Feedback.GetComment())
			comment = &value
		}
		response.Feedbacks = append(response.Feedbacks, openapi.FeedbackDetail{
			Id:        openapi.GameFeedbackID(feedback.Feedback.GetID()),
			Answers:   answers,
			Comment:   comment,
			CreatedAt: feedback.Feedback.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

func gameFeedbackPagination(limitParam, offsetParam *int) (int, int, error) {
	limit := 50
	if limitParam != nil {
		limit = *limitParam
	}
	offset := 0
	if offsetParam != nil {
		offset = *offsetParam
	}
	if limit < 1 || limit > 100 || offset < 0 {
		return 0, 0, service.ErrInvalidLimit
	}
	return limit, offset, nil
}

func feedbackAnswersToOpenAPI(answerDetails []*service.GameFeedbackAnswerDetail) ([]openapi.FeedbackAnswer, error) {
	answers := make([]openapi.FeedbackAnswer, 0, len(answerDetails))
	for _, answerDetail := range answerDetails {
		answer := openapi.FeedbackAnswer{}
		switch answerDetail.AnswerType {
		case values.FeedbackAnswerTypeYesNo:
			err := answer.FromFeedbackAnswerYesNo(openapi.FeedbackAnswerYesNo{
				QuestionID:   openapi.FeedbackQuestionID(answerDetail.Answer.GetQuestionID()),
				QuestionText: string(answerDetail.QuestionText),
				Answer:       answerDetail.Answer.GetAnswer(),
			})
			if err != nil {
				return nil, err
			}
		case values.FeedbackAnswerTypeFiveScale:
			err := answer.FromFeedbackAnswerFiveScale(openapi.FeedbackAnswerFiveScale{
				QuestionID:   openapi.FeedbackQuestionID(answerDetail.Answer.GetQuestionID()),
				QuestionText: string(answerDetail.QuestionText),
				Answer:       answerDetail.Answer.GetAnswer(),
			})
			if err != nil {
				return nil, err
			}
		default:
			return nil, service.ErrInvalidFormat
		}
		answers = append(answers, answer)
	}
	return answers, nil
}
