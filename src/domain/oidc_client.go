package domain

import "github.com/traPtitech/trap-collection-server/src/domain/values"

// OIDCClient
// OIDC・OAuth2のクライアントを表すドメイン。
// Secretは現在使用予定がないため入っていない。
type OIDCClient struct {
	clientID values.OIDCClientID
}

func NewOIDCClient(clientID values.OIDCClientID) *OIDCClient {
	return &OIDCClient{
		clientID: clientID,
	}
}

func (oc *OIDCClient) GetClientID() values.OIDCClientID {
	return oc.clientID
}
