package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVersion struct {
	gameVersionService service.GameVersionV2
	gameVersionUnimplemented
}

func NewGameVersion(gameVersionService service.GameVersionV2) *GameVersion {
	return &GameVersion{
		gameVersionService: gameVersionService,
	}
}

// gameVersionUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameVersionUnimplemented interface {
	// ゲームバージョン一覧の取得
	// (GET /games/{gameID}/versions)
	GetGameVersion(ctx echo.Context, gameID openapi.GameIDInPath, params openapi.GetGameVersionParams) error
	// ゲームのバージョンの作成
	// (POST /games/{gameID}/versions)
	PostGameVersion(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームの最新バージョンの取得
	// (GET /games/{gameID}/versions/latest)
	GetLatestGameVersion(ctx echo.Context, gameID openapi.GameIDInPath) error
}
