package v2

import (
	"fmt"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type OAuth2 struct {
	oauth2Unimplemented
	baseURL     *url.URL
	session     *Session
	oidcService service.OIDCV2
}

func NewOAuth2(conf config.Handler, session *Session, oidcService service.OIDCV2) (*OAuth2, error) {
	baseURL, err := conf.TraqBaseURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get traq base url: %v", err)
	}

	return &OAuth2{
		baseURL:     baseURL,
		session:     session,
		oidcService: oidcService,
	}, nil
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
