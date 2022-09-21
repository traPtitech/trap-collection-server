package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type GameRole struct {
	gameRoleUnimplemented
}

func NewGameRole() *GameRole {
	return &GameRole{}
}

// gameRoleUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameRoleUnimplemented interface {
	// ゲームの管理権限の変更
	// (PATCH /games/{gameID}/roles)
	PatchGameRole(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームの管理権限の削除
	// (DELETE /games/{gameID}/roles/{userID})
	DeleteGameRole(ctx echo.Context, gameID openapi.GameIDInPath, userID openapi.UserIDInPath) error
}
