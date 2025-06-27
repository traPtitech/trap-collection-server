package common

import (
	"net/http"
	"testing"

	"github.com/gorilla/sessions"
)

// New
// テスト用の関数。
func (s *Session) New(t *testing.T, req *http.Request) (*sessions.Session, error) {
	t.Helper()
	return s.store.New(req, s.key)
}

// Get
// テスト用の関数。
func (s *Session) Get(t *testing.T, req *http.Request) (*sessions.Session, error) {
	t.Helper()
	return s.store.Get(req, s.key)
}
