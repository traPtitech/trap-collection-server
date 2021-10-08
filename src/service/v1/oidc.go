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
