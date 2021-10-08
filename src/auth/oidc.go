package auth

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type OIDC interface {
	GetOIDCSession(ctx context.Context, client *domain.OIDCClient, code values.OIDCAuthorizationCode, authState *domain.OIDCAuthState) (*domain.OIDCSession, error)
	RevokeOIDCSession(ctx context.Context, client *domain.OIDCClient, session *domain.OIDCSession) error
}
