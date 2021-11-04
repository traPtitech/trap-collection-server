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
	SwiftAuthURL   *url.URL
	SwiftUserName  string
	SwiftPassword  string
	SwiftTenantID  string
	SwiftContainer string
	FilePath       string
)
