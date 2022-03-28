package service

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type OIDC interface {
	Authorize(ctx context.Context) (*domain.OIDCClient, *domain.OIDCAuthState, error)
	Callback(ctx context.Context, authState *domain.OIDCAuthState, code values.OIDCAuthorizationCode) (*domain.OIDCSession, error)
	Logout(ctx context.Context, session *domain.OIDCSession) error
	TraPAuth(ctx context.Context, session *domain.OIDCSession) error
}
