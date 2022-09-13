package v2

import (
	"errors"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
)

type Session struct {
	*common.Session
}

func NewSession(session *common.Session) (*Session, error) {
	return &Session{
		Session: session,
	}, nil
}

// getSession
// セッションを取得する
// v1実装削除後はcommon.Sessionを消しセッション取得をprivateな関数にしたいため、
// common.Session.GetSessionを直接呼ばずにここでラップしておく
//
//nolint:unused
func (s *Session) get(c echo.Context) (*sessions.Session, error) {
	return s.Session.GetSession(c)
}

// save
// セッションを保存する
// v1実装削除後はcommon.Sessionを消しセッション保存をprivateな関数にしたいため、
// common.Session.Saveを直接呼ばずにここでラップしておく
//
//nolint:unused
func (s *Session) save(c echo.Context, session *sessions.Session) error {
	return s.Session.Save(c, session)
}

// revoke
// セッションを破棄する
// v1実装削除後はcommon.Sessionを消しセッション破棄をprivateな関数にしたいため、
// common.Session.Revokeを直接呼ばずにここでラップしておく
//
//nolint:unused
func (s *Session) revoke(session *sessions.Session) {
	s.Session.Revoke(session)
}

var (
	ErrNoValue     = errors.New("no value")
	ErrValueBroken = errors.New("value broken")
)
