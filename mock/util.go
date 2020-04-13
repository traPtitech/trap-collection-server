package mock

import (
	"github.com/traPtitech/trap-collection-server/openapi"
)

// API apiの構造体（mock）
type API struct {
	User openapi.User
	openapi.Api
}