package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type OIDCSession struct {
	accessToken values.OIDCAccessToken
	expiresAt   time.Time
}

func NewOIDCSession(accessToken values.OIDCAccessToken, expiresAt time.Time) *OIDCSession {
	return &OIDCSession{
		accessToken: accessToken,
		expiresAt:   expiresAt,
	}
}

func (s *OIDCSession) GetAccessToken() values.OIDCAccessToken {
	return s.accessToken
}

func (s *OIDCSession) GetExpiresAt() time.Time {
	return s.expiresAt
}

func (s *OIDCSession) IsExpired() bool {
	return s.expiresAt.Before(time.Now())
}
