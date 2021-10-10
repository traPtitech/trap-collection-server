package v1

import (
	"encoding/gob"
	"time"

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

	/*
		gorilla/sessionsの内部で使われているgobが
		time.Timeのエンコード・デコードをできるようにRegisterする
	*/
	gob.Register(time.Time{})

	return &Session{
		key:    string(key),
		secret: string(secret),
		store:  store,
	}
}

func (s *Session) Use(e *echo.Echo) {
	e.Use(session.Middleware(s.store))
}

/*func (s *Session) getSession(c echo.Context) (*sessions.Session, error) {
	session, err := s.store.Get(c.Request(), s.key)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (s *Session) save(c echo.Context, session *sessions.Session) error {
	err := s.store.Save(c.Request(), c.Response(), session)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (s *Session) revoke(session *sessions.Session) {
	session.Options.MaxAge = -1
}

var (
	ErrNoValue     = errors.New("no value")
	ErrValueBroken = errors.New("value broken")
)

const (
	codeVerifierSessionKey = "codeVerifier"
	// 旧実装と共存させるために、accessTokenとexpiresAtを別々に保存
	accessTokenSessionKey = "accessToken"
	expiresAtSessionKey   = "expiresAt"
)

func (s *Session) setCodeVerifier(session *sessions.Session, codeVerifier string) {
	session.Values[codeVerifierSessionKey] = codeVerifier
}

func (s *Session) getCodeVerifier(session *sessions.Session) (string, error) {
	iCodeVerifier, ok := session.Values[codeVerifierSessionKey]
	if !ok {
		return "", ErrNoValue
	}

	codeVerifier, ok := iCodeVerifier.(string)
	if !ok {
		return "", ErrValueBroken
	}

	return codeVerifier, nil
}

func (s *Session) setAuthSession(session *sessions.Session, authSession *domain.OIDCSession) {
	session.Values[accessTokenSessionKey] = string(authSession.GetAccessToken())
	session.Values[expiresAtSessionKey] = authSession.GetExpiresAt()
}

func (s *Session) getAuthSession(session *sessions.Session) (*domain.OIDCSession, error) {
	iAccessToken, ok := session.Values[accessTokenSessionKey]
	if !ok {
		return nil, ErrNoValue
	}

	accessToken, ok := iAccessToken.(string)
	if !ok {
		return nil, ErrValueBroken
	}

	iExpiresAt, ok := session.Values[expiresAtSessionKey]
	if !ok {
		return nil, ErrNoValue
	}

	expiresAt, ok := iExpiresAt.(time.Time)
	if !ok {
		return nil, ErrValueBroken
	}

	return domain.NewOIDCSession(
		values.NewOIDCAccessToken(accessToken),
		expiresAt,
	), nil
}*/
