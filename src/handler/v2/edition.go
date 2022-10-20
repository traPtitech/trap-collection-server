package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Edition struct {
	editionService service.Edition
	editionUnimplemented
}

func NewEdition(editionService service.Edition) *Edition {
	return &Edition{
		editionService: editionService,
	}
}

// editionUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type editionUnimplemented interface {
	// エディション一覧の取得
	// (GET /editions)
	GetEditions(ctx echo.Context) error
	// エディションの作成
	// (POST /editions)
	PostEdition(ctx echo.Context) error
	// エディションの削除
	// (DELETE /editions/{editionID})
	DeleteEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディション情報の取得
	// (GET /editions/{editionID})
	GetEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディション情報の変更
	// (PATCH /editions/{editionID})
	PatchEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディションに紐づくゲームの一覧の取得
	// (GET /editions/{editionID}/games)
	GetEditionGames(ctx echo.Context, editionID openapi.EditionIDInPath) error
	// エディションのゲームの変更
	// (PATCH /editions/{editionID}/games)
	PostEditionGame(ctx echo.Context, editionID openapi.EditionIDInPath) error
}
