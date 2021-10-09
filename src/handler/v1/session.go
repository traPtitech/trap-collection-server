package v1

import (
	"github.com/gorilla/sessions"
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
