//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2"
)

var repositorySet = wire.NewSet(
	wire.Bind(new(repository.DB), new(*gorm2.DB)),
	gorm2.NewDB,

	wire.Bind(new(repository.Game), new(*gorm2.Game)),
	gorm2.NewGame,

	wire.Bind(new(repository.GameVersion), new(*gorm2.GameVersion)),
	gorm2.NewGameVersion,

	wire.Bind(new(repository.GameImage), new(*gorm2.GameImage)),
	gorm2.NewGameImage,

	wire.Bind(new(repository.GameVideo), new(*gorm2.GameVideo)),
	gorm2.NewGameVideo,

	wire.Bind(new(repository.GameFile), new(*gorm2.GameFile)),
	gorm2.NewGameFile,

	wire.Bind(new(repository.GameURL), new(*gorm2.GameURL)),
	gorm2.NewGameURL,

	wire.Bind(new(repository.GameManagementRole), new(*gorm2.GameManagementRole)),
	gorm2.NewGameManagementRole,

	wire.Bind(new(repository.LauncherSession), new(*gorm2.LauncherSession)),
	gorm2.NewLauncherSession,

	wire.Bind(new(repository.LauncherUser), new(*gorm2.LauncherUser)),
	gorm2.NewLauncherUser,

	wire.Bind(new(repository.LauncherVersion), new(*gorm2.LauncherVersion)),
	gorm2.NewLauncherVersion,

	wire.Bind(new(repository.GameImageV2), new(*gorm2.GameImageV2)),
	gorm2.NewGameImageV2,

	wire.Bind(new(repository.GameV2), new(*gorm2.GameV2)),
	gorm2.NewGameV2,

	wire.Bind(new(repository.GameVersionV2), new(*gorm2.GameVersionV2)),
	gorm2.NewGameVersionV2,

	wire.Bind(new(repository.GameVideoV2), new(*gorm2.GameVideoV2)),
	gorm2.NewGameVideoV2,

	wire.Bind(new(repository.GameFileV2), new(*gorm2.GameFileV2)),
	gorm2.NewGameFileV2,

	wire.Bind(new(repository.Edition), new(*gorm2.Edition)),
	gorm2.NewEdition,

	wire.Bind(new(repository.ProductKey), new(*gorm2.ProductKey)),
	gorm2.NewProductKey,

	wire.Bind(new(repository.AccessToken), new(*gorm2.AccessToken)),
	gorm2.NewAccessToken,

	wire.Bind(new(repository.AdminAuthV2), new(*gorm2.AdminAuth)),
	gorm2.NewAdminAuth,

	wire.Bind(new(repository.Seat), new(*gorm2.Seat)),
	gorm2.NewSeat,

	wire.Bind(new(repository.GameGenre), new(*gorm2.GameGenre)),
	gorm2.NewGameGenre,
)
