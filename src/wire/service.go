//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/service"
	v1 "github.com/traPtitech/trap-collection-server/src/service/v1"
	v2 "github.com/traPtitech/trap-collection-server/src/service/v2"
)

var (
	serviceSet = wire.NewSet(
		serviceV1Set,
		serviceV2Set,
	)
	serviceV1Set = wire.NewSet(
		// 	wire.Bind(new(service.AdministratorAuth), new(*v1.AdministratorAuth)),
		// 	v1.NewAdministratorAuth,

		// 	wire.Bind(new(service.GameAuth), new(*v1.GameAuth)),
		// 	v1.NewGameAuth,

		// 	wire.Bind(new(service.Game), new(*v1.Game)),
		// 	v1.NewGame,

		// 	wire.Bind(new(service.GameVersion), new(*v1.GameVersion)),
		// 	v1.NewGameVersion,

		// 	wire.Bind(new(service.GameImage), new(*v1.GameImage)),
		// 	v1.NewGameImage,

		// 	wire.Bind(new(service.GameVideo), new(*v1.GameVideo)),
		// 	v1.NewGameVideo,

		// 	wire.Bind(new(service.GameFile), new(*v1.GameFile)),
		// 	v1.NewGameFile,

		// 	wire.Bind(new(service.GameURL), new(*v1.GameURL)),
		// 	v1.NewGameURL,

		// 	wire.Bind(new(service.LauncherAuth), new(*v1.LauncherAuth)),
		// 	v1.NewLauncherAuth,

		// 	wire.Bind(new(service.LauncherVersion), new(*v1.LauncherVersion)),
		// 	v1.NewLauncherVersion,

		// 	wire.Bind(new(service.OIDC), new(*v1.OIDC)),
		// 	v1.NewOIDC,

		wire.Bind(new(service.User), new(*v1.User)),
		v1.NewUser,

		v1.NewUserUtils,
	)
	serviceV2Set = wire.NewSet(
		wire.Bind(new(service.OIDCV2), new(*v2.OIDC)),
		v2.NewOIDC,

		wire.Bind(new(service.GameImageV2), new(*v2.GameImage)),
		v2.NewGameImage,

		wire.Bind(new(service.GameV2), new(*v2.Game)),
		v2.NewGame,

		wire.Bind(new(service.GameVideoV2), new(*v2.GameVideo)),
		v2.NewGameVideo,

		wire.Bind(new(service.GameVersionV2), new(*v2.GameVersion)),
		v2.NewGameVersion,

		wire.Bind(new(service.GameFileV2), new(*v2.GameFile)),
		v2.NewGameFile,

		wire.Bind(new(service.Edition), new(*v2.Edition)),
		v2.NewEdition,

		wire.Bind(new(service.EditionAuth), new(*v2.EditionAuth)),
		v2.NewEditionAuth,

		wire.Bind(new(service.GameRoleV2), new(*v2.GameRole)),
		v2.NewGameRole,

		wire.Bind(new(service.AdminAuthV2), new(*v2.AdminAuth)),
		v2.NewAdminAuth,

		wire.Bind(new(service.Seat), new(*v2.Seat)),
		v2.NewSeat,

		wire.Bind(new(service.GameGenre), new(*v2.GameGenre)),
		v2.NewGameGenre,

		wire.Bind(new(service.GamePlayLogV2), new(*v2.GamePlayLog)),
		v2.NewGamePlayLog,

		// wire.Bind(new(service.User), new(*v1.User)),
		// v1.NewUser,

		v2.NewUser,
	)
)
