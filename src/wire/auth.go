//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/auth"
	traq "github.com/traPtitech/trap-collection-server/src/auth/traQ"
)

var authSet = wire.NewSet(
	wire.Bind(new(auth.OIDC), new(*traq.OIDC)),
	traq.NewOIDC,

	wire.Bind(new(auth.User), new(*traq.User)),
	traq.NewUser,
)
