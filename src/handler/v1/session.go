package v1

import (
	"encoding/gob"
	"errors"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
)

type Session struct {
	*common.Session
}

func NewSession(session *common.Session) (*Session, error) {
	// gorilla/sessionsの内部で使われているgobが
	// time.Timeのエンコード・デコードをできるようにRegisterする
	gob.Register(time.Time{})

	return &Session{
		Session: session,
	}, nil
}

func (s *Session) getSession(c echo.Context) (*sessions.Session, error) {
	return s.Session.GetSession(c)
}

func (s *Session) save(c echo.Context, session *sessions.Session) error {
	return s.Session.Save(c, session)
}

func (s *Session) revoke(session *sessions.Session) {
	s.Session.Revoke(session)
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
}
