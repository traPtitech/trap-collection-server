package v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type OIDC struct {
	client   *domain.OIDCClient
	oidcAuth auth.OIDC
}

func NewOIDC(conf config.ServiceV1, oidc auth.OIDC) (*OIDC, error) {
	strClientID, err := conf.ClientID()
	if err != nil {
		return nil, fmt.Errorf("failed to get client ID: %w", err)
	}

	clientID := values.NewOIDCClientID(string(strClientID))

	client := domain.NewOIDCClient(clientID)

	return &OIDC{
		client:   client,
		oidcAuth: oidc,
	}, nil
}

func (o *OIDC) Authorize(ctx context.Context) (*domain.OIDCClient, *domain.OIDCAuthState, error) {
	codeChallengeMethod := values.OIDCCodeChallengeMethodSha256
	codeChallenge, err := values.NewOIDCCodeVerifier()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	state := domain.NewOIDCAuthState(codeChallengeMethod, codeChallenge)

	return o.client, state, nil
}

func (o *OIDC) Callback(ctx context.Context, authState *domain.OIDCAuthState, code values.OIDCAuthorizationCode) (*domain.OIDCSession, error) {
	session, err := o.oidcAuth.GetOIDCSession(ctx, o.client, code, authState)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return nil, service.ErrInvalidAuthStateOrCode
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC session: %w", err)
	}

	return session, nil
}

func (o *OIDC) Logout(ctx context.Context, session *domain.OIDCSession) error {
	err := o.oidcAuth.RevokeOIDCSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to revoke OIDC session: %w", err)
	}

	return nil
}

// traQで凍結された場合の反映が遅れるのは許容しているので、sessionの有効期限確認のみ
func (o *OIDC) TraPAuth(ctx context.Context, session *domain.OIDCSession) error {
	if session.IsExpired() {
		return service.ErrOIDCSessionExpired
	}

	return nil
}
