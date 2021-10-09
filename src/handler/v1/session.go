package v1

import (
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/pkg/common"
)

type Session struct {
	key    string
	secret string
	store  sessions.Store
}

func NewSession(key common.SessionKey, secret common.SessionSecret) *Session {
	store := sessions.NewCookieStore([]byte(secret))

	return &Session{
		key:    string(key),
		secret: string(secret),
		store:  store,
	}
}

func (s *Session) Use(e *echo.Echo) {
	e.Use(session.Middleware(s.store))
}

func (s *Session) getSession(c echo.Context) (*sessions.Session, error) {
	session, err := s.store.Get(c.Request(), s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}
