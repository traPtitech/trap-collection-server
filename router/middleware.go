package router

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
)

// User ユーザーの構造体
type User struct {
	ID   string `json:"userId,omitempty"`
	Name string `json:"name,omitempty"`
}

// Traq traQのOAuthのClient
type Traq interface {
	GetMe(c echo.Context) (User, error)
	MiddlewareAuthUser(next echo.HandlerFunc) echo.HandlerFunc
}

// TraqClient 本番用のclient
type TraqClient struct {
	Traq
}

// MockTraqClient テスト用のモックclient
type MockTraqClient struct {
	Traq
	User User
}

// MiddlewareAuthLancher ランチャーの認証用のミドルウェア
func MiddlewareAuthLancher(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.Request().Header.Get("X-Key")
		isThere := model.CheckProductKey(key)
		if !isThere {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}
