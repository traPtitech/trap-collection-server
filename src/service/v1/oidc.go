package v1

import (
	"context"
	"fmt"

	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type OIDC struct {
	client   *domain.OIDCClient
	oidcAuth auth.OIDC
}

func NewOIDC(oidc auth.OIDC, strClientID common.ClientID) *OIDC {
	clientID := values.NewOIDCClientID(string(strClientID))

	client := domain.NewOIDCClient(clientID)

	return &OIDC{
		client:   client,
		oidcAuth: oidc,
	}
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
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC session: %w", err)
	}

	return session, nil
}
