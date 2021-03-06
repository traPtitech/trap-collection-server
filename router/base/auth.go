package base

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// OAuth OAuthの認証の基本部分の構造体
type OAuth interface {
	BaseURL() *url.URL
	GetMe(accessToken string) (user *openapi.User, err error)
	GetUsers(accessToken string) (user []*openapi.User, err error)
}

// LauncherAuth ランチャーの認証の基本部分の構造体
type LauncherAuth interface {
	GetVersionID(c echo.Context) (versionID string, err error)
	GetProductKey(c echo.Context) (productKey string, err error)
}
