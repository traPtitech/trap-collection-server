package v2

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type GameCreator struct{}

// ゲームクリエイターのジョブ一覧の取得
// (GET /creators/jobs)
func (gc *GameCreator) GetGameCreatorJobs(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームクリエイター一覧の取得
// (GET /games/{gameID}/creators)
func (gc *GameCreator) GetGameCreators(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームクリエイター一覧の更新
// (PATCH /games/{gameID}/creators)
func (gc *GameCreator) PatchGameCreators(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームクリエイターの削除
// (DELETE /games/{gameID}/creators/{userID})
func (gc *GameCreator) DeleteGameCreator(c echo.Context, _ openapi.GameIDInPath, _ openapi.UserIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}
