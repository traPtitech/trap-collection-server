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
// (POST /game-play-logs/start)
func (gpl *GamePlayLog) PostGamePlayLogStart(ctx echo.Context) error {
	// TODO: 実装が必要
	return echo.NewHTTPError(http.StatusNotImplemented, "not implemented yet")
}

// ゲーム終了ログの記録
// (PATCH /game-play-logs/{playLogID}/end)
func (gpl *GamePlayLog) PatchGamePlayLogEnd(ctx echo.Context, playLogID openapi.PlayLogIDInPath) error {
	// TODO: 実装が必要
	return echo.NewHTTPError(http.StatusNotImplemented, "not implemented yet")
}
