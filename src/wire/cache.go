//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/cache/ristretto"
)

var cacheSet = wire.NewSet(
	wire.Bind(new(cache.User), new(*ristretto.User)),
	ristretto.NewUser,
)
