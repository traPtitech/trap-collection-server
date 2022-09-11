package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type GameVideo struct {
	gameVideoUnimplemented
}

func NewGameVideo() *GameVideo {
	return &GameVideo{}
}

// gameVideoUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameVideoUnimplemented interface {
	// ゲーム動画の作成
	// (GET /games/{gameID}/videos)
	GetGameVideos(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム動画一覧の取得
	// (POST /games/{gameID}/videos)
	PostGameVideo(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム動画のバイナリの取得
	// (GET /games/{gameID}/videos/{gameVideoID})
	GetGameVideo(ctx echo.Context, gameID openapi.GameIDInPath, gameVideoID openapi.GameVideoIDInPath) error
	// ゲーム動画のメタ情報の取得
	// (GET /games/{gameID}/videos/{gameVideoID}/meta)
	GetGameVideoMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameVideoID openapi.GameVideoIDInPath) error
}
