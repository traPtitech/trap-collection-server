package common

import (
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/config"
)

const (
	sessionContextKey = "session"
)

type Session struct {
	key    string
	secret string
	store  sessions.Store
}

func NewSession(conf config.Handler) (*Session, error) {
	secret, err := conf.SessionSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to get session secret: %w", err)
	}

	key, err := conf.SessionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get session key: %w", err)
	}

	store := sessions.NewCookieStore([]byte(secret))

	return &Session{
		key:    key,
		secret: secret,
		store:  store,
	}, nil
}

func (s *Session) Use(e *echo.Echo) {
	e.Use(session.Middleware(s.store))
}

func (s *Session) GetSession(c echo.Context) (*sessions.Session, error) {
	session, err := s.store.Get(c.Request(), s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	c.Set(sessionContextKey, session)

	return session, nil
}

func (s *Session) Save(c echo.Context, session *sessions.Session) error {
	err := s.store.Save(c.Request(), c.Response(), session)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (s *Session) Revoke(session *sessions.Session) {
	session.Options.MaxAge = -1
}
