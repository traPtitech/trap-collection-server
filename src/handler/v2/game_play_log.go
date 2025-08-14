package v2

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type GamePlayLog struct{}

func NewGamePlayLog() *GamePlayLog {
	return &GamePlayLog{}
}

// ゲーム起動ログの記録
// (POST /editions/{editionID}/games/{gameID}/plays/start)
func (gpl *GamePlayLog) PostGamePlayLogStart(_ echo.Context, _ openapi.EditionIDInPath, _ openapi.GameIDInPath) error {
	// TODO: 実装が必要
	return echo.NewHTTPError(http.StatusNotImplemented, "not implemented yet")
}

// ゲーム終了ログの記録
// (PATCH /editions/{editionID}/games/{gameID}/plays/{playLogID}/end)
func (gpl *GamePlayLog) PatchGamePlayLogEnd(_ echo.Context, _ openapi.EditionIDInPath, _ openapi.GameIDInPath, _ openapi.PlayLogIDInPath) error {
	// TODO: 実装が必要
	return echo.NewHTTPError(http.StatusNotImplemented, "not implemented yet")
}

// ゲームプレイ統計の取得
// (GET /games/{gameID}/play-stats)
func (gpl *GamePlayLog) GetGamePlayStats(_ echo.Context, _ openapi.GameIDInPath, _ openapi.GetGamePlayStatsParams) error {
	// TODO: 実装が必要
	return echo.NewHTTPError(http.StatusNotImplemented, "not implemented yet")
}

// エディションプレイ統計の取得
// (GET /editions/{editionID}/play-stats)
func (gpl *GamePlayLog) GetEditionPlayStats(_ echo.Context, _ openapi.EditionIDInPath, _ openapi.GetEditionPlayStatsParams) error {
	// TODO: 実装が必要
	return echo.NewHTTPError(http.StatusNotImplemented, "not implemented yet")
}
