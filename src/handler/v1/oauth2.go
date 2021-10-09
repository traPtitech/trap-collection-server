package v1

import (
	"github.com/traPtitech/trap-collection-server/src/service"
)

type OAuth2 struct {
	session     *Session
	oidcService service.OIDC
}

func NewOAuth2(session *Session, oidcService service.OIDC) *OAuth2 {
	return &OAuth2{
		session:     session,
		oidcService: oidcService,
	}
}
