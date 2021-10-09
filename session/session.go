package session

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

// Session セッションの構造体
type Session interface {
	Store() sessions.Store
	RevokeSession(c echo.Context) error
}

type sess struct {
	store sessions.Store
}

// NewSession Sessionのコンストラクタ
func NewSession(secret string) (Session, error) {
	newSessions := new(sess)
	store := sessions.NewCookieStore([]byte(secret))

	newSessions.store = store

	return newSessions, nil
}

func (s *sess) Store() sessions.Store {
	return s.store
}

func (s *sess) RevokeSession(c echo.Context) error {
	cookie, err := c.Cookie("sessions")
	if errors.Is(err, http.ErrNoCookie) {
		return err
	}

	cookie.MaxAge = -1
	c.SetCookie(cookie)

	return nil
}
