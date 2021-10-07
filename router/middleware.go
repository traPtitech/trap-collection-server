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
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
)

// Middleware middlewareの構造体
type Middleware struct {
	db    model.DBMeta
	oauth base.OAuth
	*v1.Middleware
}

func newMiddleware(db model.DBMeta, oauth base.OAuth, newMiddleware *v1.Middleware) openapi.Middleware {
	middleware := new(Middleware)

	middleware.db = db
	middleware.oauth = oauth
	middleware.Middleware = newMiddleware

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
		interfaceUserName, ok1 := sess.Values["userName"]
		userName, ok2 := interfaceUserName.(string)
		if !ok2 {
			log.Printf("error: unexcepted invalid userName")
			return echo.NewHTTPError(http.StatusInternalServerError, errors.New("unexpected invalid userName"))
		}
		interfaceAccessToken, ok3 := sess.Values["accessToken"]
		if !ok1 || !ok2 {
			log.Printf("error: unexcepted no userName")

			if !ok3 || interfaceAccessToken == nil {
				return c.String(http.StatusUnauthorized, errors.New("No Access Token").Error())
			}

			accessToken, ok := interfaceAccessToken.(string)
			if !ok {
				log.Printf("error: unexcepted invalid accessToken")
				return echo.NewHTTPError(http.StatusInternalServerError, errors.New("unexpected invalid accessToken"))
			}

			user, err := m.oauth.GetMe(accessToken)
			if err != nil {
				return c.String(http.StatusBadRequest, fmt.Errorf("Failed In Getting User: %w", err).Error())
			}

			userName = user.Name
		}

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

// BothAuthMiddleware ランチャー・traQの認証用のミドルウェア
func (m *Middleware) BothAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ok, err := m.CheckLauncherAuth(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		if ok {
			return next(c)
		}

		sess, err := session.Get("sessions", c)
		if err != nil {
			log.Printf("error: unexcepted no session: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("failed to get session: %v", err))
		}

		accessToken, ok := sess.Values["accessToken"]
		if !ok || accessToken == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "No Access Token")
		}

		c.Set("accessToken", accessToken)

		return next(c)
	}
}
