package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type GameGenre struct {
	gameGenreUnimplemented
}

func NewGameGenre() *GameGenre {
	return &GameGenre{}
}

// gameGenreUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameGenreUnimplemented interface {
	// 全てのジャンルの取得
	// (GET /genres)
	GetGameGenres(ctx echo.Context) error
	// ジャンルの削除
	// (DELETE /genres/{gameGenreID})
	DeleteGameGenre(ctx echo.Context, gameGenreID openapi.GameGenreIDInPath) error
	// ジャンル情報の変更
	// (PATCH /genres/{gameGenreID})
	PatchGameGenre(ctx echo.Context, gameGenreID openapi.GameGenreIDInPath) error
}
