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
	"github.com/traPtitech/trap-collection-server/router/base"
)

// Middleware middlewareの構造体
type Middleware struct {
	db    model.DBMeta
	oauth base.OAuth
	openapi.Middleware
}

func newMiddleware(db model.DBMeta, oauth base.OAuth) openapi.Middleware {
	middleware := new(Middleware)

	middleware.db = db
	middleware.oauth = oauth

	return middleware
}

// TrapMemberAuthMiddleware traQのOAuthのmiddleware
func (m *Middleware) TrapMemberAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}

		accessToken, ok := sess.Values["accessToken"]
		if !ok || accessToken == nil {
			return c.String(http.StatusUnauthorized, errors.New("No Access Token").Error())
		}

		c.Set("accessToken", accessToken)

		return next(c)
	}
}

// GameMaintainerAuthMiddleware ゲーム管理者の認証用のミドルウェア
func (m *Middleware) GameMaintainerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}

		var userID string
		var accessToken string
		interfaceUserID, ok := sess.Values["userID"]
		if !ok || interfaceUserID == nil {
			log.Println("error: unexcepted no userID")
			interfaceAccessToken, ok := sess.Values["accessToken"]
			if !ok || interfaceAccessToken == nil {
				return c.String(http.StatusUnauthorized, "No Access Token")
			}
			accessToken = interfaceAccessToken.(string)
			user, err := m.oauth.GetMe(accessToken)
			if err != nil {
				return c.String(http.StatusBadRequest, fmt.Errorf("Failed In Getting User: %w", err).Error())
			}
			userID = user.Id
		} else {
			userID = interfaceUserID.(string)
		}

		gameID := c.Param("gameID")
		if len(gameID) == 0 {
			log.Println("error: unexpected no gameID")
			return c.String(http.StatusInternalServerError, "No GameID")
		}

		isThereGame, err := m.db.IsExistGame(gameID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to check if there is the game: %w", err))
		}
		if !isThereGame {
			return echo.NewHTTPError(http.StatusNotFound, errors.New("gameID doesn't exist"))
		}

		isMaintainer, err := m.db.CheckMaintainerID(userID, gameID)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Checking MaintainerID: %w", err).Error())
		}
		if !isMaintainer {
			return c.String(http.StatusUnauthorized, "You Are Not Maintainer")
		}

		c.Set("userID", userID)
		c.Set("accessToken", accessToken)

		return next(c)
	}
}

// GameOwnerAuthMiddleware ゲーム管理者の認証用のミドルウェア
func (*Middleware) GameOwnerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

// AdminAuthMiddleware ランチャーの管理者の認証用のミドルウェア
func (m *Middleware) AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}

		// 暫定的な実装。最終的にはDBにあるAdminと比べ、userIDを使い認証するようにする。
		admins := []string{"mazrean"}
		var userName string
		var accessToken string
		interfaceUserName, ok1 := sess.Values["userName"]
		interfaceAccessToken, ok2 := sess.Values["accessToken"]
		if !ok1 || interfaceUserName == nil {
			log.Printf("error: unexcepted no userName")
			if !ok2 || interfaceAccessToken == nil {
				return c.String(http.StatusUnauthorized, errors.New("No Access Token").Error())
			}
			accessToken = interfaceAccessToken.(string)
			user, err := m.oauth.GetMe(accessToken)
			if err != nil {
				return c.String(http.StatusBadRequest, fmt.Errorf("Failed In Getting User: %w", err).Error())
			}
			userName = user.Name
		}
		userName = interfaceUserName.(string)

		for _, v := range admins {
			if v == userName {
				c.Set("userName", interfaceUserName)
				c.Set("accessToken", interfaceAccessToken)

				return next(c)
			}
		}

		return c.String(http.StatusUnauthorized, errors.New("You Are Not Admin").Error())
	}
}

// LauncherAuthMiddleware ランチャーの認証用のミドルウェア
func (m *Middleware) LauncherAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}

		interfaceProductKey := sess.Values["productKey"]
		if interfaceProductKey != nil {
			productKey, ok := interfaceProductKey.(string)
			if ok {
				isThere, versionID := m.db.CheckProductKey(productKey)
				if isThere {
					log.Printf("debug: %d", versionID)
					sess.Values["versionID"] = versionID
					sess.Save(c.Request(), c.Response())

					return next(c)
				}
			}
		}

		key := c.Request().Header.Get("X-Key")
		isThere, versionID := m.db.CheckProductKey(key)
		if !isThere {
			return c.NoContent(http.StatusUnauthorized)
		}
		log.Printf("debug: %d", versionID)

		sess.Values["productKey"] = key
		sess.Values["version_id"] = versionID
		sess.Save(c.Request(), c.Response())

		return next(c)
	}
}

// BothAuthMiddleware ランチャー・traQの認証用のミドルウェア
func (*Middleware) BothAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
