package session

import (
	"database/sql"
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/srinathgs/mysqlstore"
)

// Session セッションの構造体
type Session interface {
	Store() sessions.Store
	RevokeSession(c echo.Context) error
}

type sess struct {
	store *mysqlstore.MySQLStore
}

// NewSession Sessionのコンストラクタ
func NewSession(db *sql.DB) (Session, error) {
	newSessions := new(sess)
	store, err := mysqlstore.NewMySQLStoreFromConnection(db, "sessions", "/", 60*60*24*14, []byte("secret-token"))
	if err != nil {
		return &sess{}, fmt.Errorf("Failed In Creating Store: %w", err)
	}

	newSessions.store = store

	return newSessions, nil
}

func (s *sess) Store() sessions.Store {
	return s.store
}

func (s *sess) RevokeSession(c echo.Context) error {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return fmt.Errorf("Failed In Getting Session: %w", err)
	}

	err = s.store.Delete(c.Request(), c.Response(), sess)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
