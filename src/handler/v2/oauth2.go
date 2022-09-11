package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type OAuth2 struct {
	oauth2Unimplemented
}

func NewOAuth2() *OAuth2 {
	return &OAuth2{}
}

// oauth2Unimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type oauth2Unimplemented interface {
	// traQのOAuth 2.0のコールバック
	// (GET /oauth2/callback)
	GetCallback(ctx echo.Context, params openapi.GetCallbackParams) error
	// OAuth 2.0のCode Verifierなどのセッションへの設定とtraQへのリダイレクト
	// (GET /oauth2/code)
	GetCode(ctx echo.Context) error
	// traP Collectionの管理画面からのログアウト
	// (POST /oauth2/logout)
	PostLogout(ctx echo.Context) error
}
