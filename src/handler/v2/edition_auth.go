package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type EditionAuth struct {
	editionAuthUnimplemented
}

func NewEditionAuth() *EditionAuth {
	return &EditionAuth{}
}

// editionAuthUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type editionAuthUnimplemented interface {
	// プロダクトキーの一覧の取得
	// (GET /editions/{editionID}/keys)
	GetProductKeys(ctx echo.Context, editionID openapi.EditionIDInPath, params openapi.GetProductKeysParams) error
	// プロダクトキーの生成
	// (POST /editions/{editionID}/keys)
	PostProductKey(ctx echo.Context, editionID openapi.EditionIDInPath, params openapi.PostProductKeyParams) error
	// プロダクトキーの再有効化
	// (POST /editions/{editionID}/keys/{productKeyID}/activate)
	PostActivateProductKey(ctx echo.Context, editionID openapi.EditionIDInPath, productKeyID openapi.ProductKeyIDInPath) error
	// プロダクトキーの失効
	// (POST /editions/{editionID}/keys/{productKeyID}/revoke)
	PostRevokeProductKey(ctx echo.Context, editionID openapi.EditionIDInPath, productKeyID openapi.ProductKeyIDInPath) error
	// ランチャーの認可リクエスト
	// (POST /editions/authorize)
	PostEditionAuthorize(ctx echo.Context) error
	// エディション情報の取得
	// (GET /editions/info)
	GetEditionInfo(ctx echo.Context) error
}
