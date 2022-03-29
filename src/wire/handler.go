//go:build wireinject

package wire

import (
	"github.com/google/wire"
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
)

var handlerSet = wire.NewSet(
	v1.NewAPI,
	v1.NewSession,
	v1.NewGame,
	v1.NewGameRole,
	v1.NewGameImage,
	v1.NewGameVideo,
	v1.NewGameVersion,
	v1.NewGameFile,
	v1.NewGameURL,
	v1.NewLauncherAuth,
	v1.NewLauncherVersion,
	v1.NewOAuth2,
	v1.NewUser,
	v1.NewMiddleware,
)
