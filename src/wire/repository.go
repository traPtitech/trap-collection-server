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
)