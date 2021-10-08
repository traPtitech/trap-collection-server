package v1

import (
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
