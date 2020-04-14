package router

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Middleware middlewareの構造体
type Middleware struct {
	openapi.Middleware
}

// BasicMiddleware ランチャーの認証用のミドルウェア
func (m Middleware) BasicMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

// OAuthMiddleware traQのOAuthのmiddleware
func (m Middleware) OAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}
		accessToken := sess.Values["accessToken"]
		if accessToken == nil {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}

// BothMiddleware ランチャーの認証用のミドルウェア
func (m Middleware) BothMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
