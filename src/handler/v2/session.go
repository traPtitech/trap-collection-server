package v2

import (
	"encoding/gob"
	"errors"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/session"
)

type Session struct {
	*session.Session
}

func NewSession(session *session.Session) (*Session, error) {
	// gorilla/sessionsの内部で使われているgobが
	// time.Timeのエンコード・デコードをできるようにRegisterする
	gob.Register(time.Time{})

	return &Session{
		Session: session,
	}, nil
}

// getSession
// セッションを取得する
// v1実装削除後はsession.Sessionを消しセッション取得をprivateな関数にしたいため、
// session.Session.GetSessionを直接呼ばずにここでラップしておく
//
//nolint:unused
func (s *Session) get(c echo.Context) (*sessions.Session, error) {
	return s.GetSession(c)
}

// save
// セッションを保存する
// v1実装削除後はsession.Sessionを消しセッション保存をprivateな関数にしたいため、
// session.Session.Saveを直接呼ばずにここでラップしておく
//
//nolint:unused
func (s *Session) save(c echo.Context, session *sessions.Session) error {
	return s.Save(c, session)
}

// revoke
// セッションを破棄する
// v1実装削除後はsession.Sessionを消しセッション破棄をprivateな関数にしたいため、
// session.Session.Revokeを直接呼ばずにここでラップしておく
//
//nolint:unused
func (s *Session) revoke(session *sessions.Session) {
	s.Revoke(session)
}

var (
	ErrNoValue     = errors.New("no value")
	ErrValueBroken = errors.New("value broken")
)

const (
	codeVerifierSessionKey = "codeVerifier"
	accessTokenSessionKey  = "accessToken"
	expiresAtSessionKey    = "expiresAt"
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
