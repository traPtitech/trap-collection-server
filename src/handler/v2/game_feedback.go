package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
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
func (gf *GameFeedback) GetFeedbackQuestions(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// フィードバック質問の一括設定
// (PUT /games/{gameID}/feedback-questions)
func (gf *GameFeedback) PutFeedbackQuestions(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
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
