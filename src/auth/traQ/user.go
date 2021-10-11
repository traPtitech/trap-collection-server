package traq

import (
	"net/http"
	"net/url"

	"github.com/traPtitech/trap-collection-server/pkg/common"
)

type User struct {
	client  *http.Client
	baseURL *url.URL
}

func NewUser(client *http.Client, baseURL common.TraQBaseURL) *User {
	return &User{
		client:  client,
		baseURL: (*url.URL)(baseURL),
	}
}
