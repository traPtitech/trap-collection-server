package v2

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type GameFeedback struct{}

// フィードバック設定の取得
// (GET /games/{gameID}/feedback-config)
func (gf *GameFeedback) GetFeedbackConfig(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
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
