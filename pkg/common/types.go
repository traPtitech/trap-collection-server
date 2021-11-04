package common

import (
	"net/url"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type (
	IsProduction   bool
	ClientID       string
	TraQBaseURL    *url.URL
	SessionSecret  string
	SessionKey     string
	Administrators []values.TraPMemberName
	FilePath       string
)
