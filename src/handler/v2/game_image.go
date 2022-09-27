package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameImage struct {
	gameImageUnimplemented
	gameImageService service.GameImageV2
}

func NewGameImage(gameImageService service.GameImageV2) *GameImage {
	return &GameImage{
		gameImageService: gameImageService,
	}
}

// gameImageUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameImageUnimplemented interface {
	// ゲーム画像の作成
	// (GET /games/{gameID}/images)
	GetGameImages(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム画像一覧の取得
	// (POST /games/{gameID}/images)
	PostGameImage(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム画像のバイナリの取得
	// (GET /games/{gameID}/images/{gameImageID})
	GetGameImage(ctx echo.Context, gameID openapi.GameIDInPath, gameImageID openapi.GameImageIDInPath) error
	// ゲーム画像のメタ情報の取得
	// (GET /games/{gameID}/images/{gameImageID}/meta)
	GetGameImageMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameImageID openapi.GameImageIDInPath) error
}
