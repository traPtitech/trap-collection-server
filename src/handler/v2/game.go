package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type Game struct {
	gameUnimplemented
}

func NewGame() *Game {
	return &Game{}
}

// gameUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameUnimplemented interface {
	// ゲーム一覧の取得
	// (GET /games)
	GetGames(ctx echo.Context, params openapi.GetGamesParams) error
	// ゲームの追加
	// (POST /games)
	PostGame(ctx echo.Context) error
	// ゲームの削除
	// (DELETE /games/{gameID})
	DeleteGame(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム情報の取得
	// (GET /games/{gameID})
	GetGame(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームの情報の変更
	// (PATCH /games/{gameID})
	PatchGame(ctx echo.Context, gameID openapi.GameIDInPath) error
}
