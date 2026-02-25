package v2

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type GameFeedback struct{}

// フィードバック質問一覧の取得
// (GET /feedbacks/questions)
func (gf *GameFeedback) GetFeedbackQuestions(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームフィードバックの送信
// (POST /editions/{editionID}/games/{gameID}/feedbacks)
func (gf *GameFeedback) PostGameFeedback(c echo.Context, _ openapi.EditionIDInPath, _ openapi.GameIDInPath) error {
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
