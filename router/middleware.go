package router

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Middleware middlewareの構造体
type Middleware struct {
	openapi.Middleware
}

// TrapMemberAuthMiddleware traQのOAuthのmiddleware
func (m Middleware) TrapMemberAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}
		accessToken, ok := sess.Values["accessToken"]
		if !ok || accessToken == nil {
			return c.String(http.StatusUnauthorized, errors.New("No Access Token").Error())
		}
		return next(c)
	}
}

// GameMaintainerAuthMiddleware ゲーム管理者の認証用のミドルウェア
func (m Middleware) GameMaintainerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}

		var userID string
		interfaceUserID, ok := sess.Values["userID"]
		if !ok || interfaceUserID == nil {
			log.Println("error: unexcepted no userID")
			accessToken, ok := sess.Values["accessToken"]
			if !ok || accessToken == nil {
				return c.String(http.StatusUnauthorized, "No Access Token")
			}
			user, err := getMe(accessToken.(string))
			if err != nil {
				return c.String(http.StatusBadRequest, fmt.Errorf("Failed In Getting User: %w", err).Error())
			}
			userID = user.UserId
		} else {
			userID = interfaceUserID.(string)
		}

		gameID := c.Param("gameID")
		if len(gameID) == 0 {
			log.Println("error: unexpected no gameID")
			return c.String(http.StatusInternalServerError, "No GameID")
		}

		isMaintainer, err := model.CheckMaintainerID(userID, gameID)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Checking MaintainerID: %w", err).Error())
		}
		if !isMaintainer {
			return c.String(http.StatusUnauthorized, "You Are Not Maintainer")
		}

		return next(c)
	}
}

// GameOwnerAuthMiddleware ゲーム管理者の認証用のミドルウェア
func (m Middleware) GameOwnerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

// AdminAuthMiddleware ランチャーの管理者の認証用のミドルウェア
func (m Middleware) AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}

		// 暫定的な実装。最終的にはDBにあるAdminと比べ、userIDを使い認証するようにする。
		admins := []string{"mazrean"}
		userName, ok := sess.Values["userName"]
		if !ok || userName == nil {
			log.Printf("error: unexcepted no userName")
			accessToken, ok := sess.Values["accessToken"]
			if !ok || accessToken == nil {
				return c.String(http.StatusUnauthorized, errors.New("No Access Token").Error())
			}
			user, err := getMe(accessToken.(string))
			if err != nil {
				return c.String(http.StatusBadRequest, fmt.Errorf("Failed In Getting User: %w", err).Error())
			}
			userName = user.Name
		}

		for _,v := range admins {
			if v== userName {
				return next(c)
			}
		}

		return c.String(http.StatusUnauthorized, errors.New("You Are Not Admin").Error())
	}
}

// LauncherAuthMiddleware ランチャーの認証用のミドルウェア
func (m Middleware) LauncherAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		key := c.Request().Header.Get("X-Key")
		isThere := model.CheckProductKey(key)
		if !isThere {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}

// BothAuthMiddleware ランチャー・traQの認証用のミドルウェア
func (m Middleware) BothAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

