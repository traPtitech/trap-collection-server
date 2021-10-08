package traq

import (
	"net/http"
	"net/url"

	"github.com/traPtitech/trap-collection-server/pkg/common"
)

type OIDC struct {
	client  *http.Client
	baseURL *url.URL
}

func NewOIDC(client *http.Client, baseURL common.TraQBaseURL) *OIDC {
	return &OIDC{
		client:  client,
		baseURL: (*url.URL)(baseURL),
	}
}
