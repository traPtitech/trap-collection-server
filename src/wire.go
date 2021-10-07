//go:build wireinject
// +build wireinject

package src

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	v1Handler "github.com/traPtitech/trap-collection-server/src/handler/v1"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2"
	"github.com/traPtitech/trap-collection-server/src/service"
	v1Service "github.com/traPtitech/trap-collection-server/src/service/v1"
)

var (
	dbBind                        = wire.Bind(new(repository.DB), new(*gorm2.DB))
	launcherSessionRepositoryBind = wire.Bind(new(repository.LauncherSession), new(*gorm2.LauncherSession))
	launcherUserRepositoryBind    = wire.Bind(new(repository.LauncherUser), new(*gorm2.LauncherUser))
	launcherVersionRepositoryBind = wire.Bind(new(repository.LauncherVersion), new(*gorm2.LauncherVersion))

	launcherAuthServiceBind = wire.Bind(new(service.LauncherAuth), new(*v1Service.LauncherAuth))
)

func InjectAPI(isProduction common.IsProduction) (*v1Handler.API, error) {
	wire.Build(
		dbBind,
		launcherSessionRepositoryBind,
		launcherUserRepositoryBind,
		launcherVersionRepositoryBind,
		launcherAuthServiceBind,
		gorm2.NewDB,
		gorm2.NewLauncherSession,
		gorm2.NewLauncherUser,
		gorm2.NewLauncherVersion,
		v1Service.NewLauncherAuth,
		v1Handler.NewAPI,
		v1Handler.NewLauncherAuth,
		v1Handler.NewMiddleware,
	)
	return nil, nil
}
