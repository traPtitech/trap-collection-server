//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/service"
	v1 "github.com/traPtitech/trap-collection-server/src/service/v1"
)

var serviceSet = wire.NewSet(
	wire.Bind(new(service.AdministratorAuth), new(*v1.AdministratorAuth)),
	v1.NewAdministratorAuth,

	wire.Bind(new(service.GameAuth), new(*v1.GameAuth)),
	v1.NewGameAuth,

	wire.Bind(new(service.Game), new(*v1.Game)),
	v1.NewGame,

	wire.Bind(new(service.GameVersion), new(*v1.GameVersion)),
	v1.NewGameVersion,

	wire.Bind(new(service.GameImage), new(*v1.GameImage)),
	v1.NewGameImage,

	wire.Bind(new(service.GameVideo), new(*v1.GameVideo)),
	v1.NewGameVideo,

	wire.Bind(new(service.GameFile), new(*v1.GameFile)),
	v1.NewGameFile,

	wire.Bind(new(service.GameURL), new(*v1.GameURL)),
	v1.NewGameURL,

	wire.Bind(new(service.LauncherAuth), new(*v1.LauncherAuth)),
	v1.NewLauncherAuth,

	wire.Bind(new(service.LauncherVersion), new(*v1.LauncherVersion)),
	v1.NewLauncherVersion,

	wire.Bind(new(service.OIDC), new(*v1.OIDC)),
	v1.NewOIDC,

	wire.Bind(new(service.User), new(*v1.User)),
	v1.NewUser,

	v1.NewUserUtils,
)