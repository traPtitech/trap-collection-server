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

type Config struct {
	IsProduction  common.IsProduction
	SessionKey    common.SessionKey
	SessionSecret common.SessionSecret
}

var (
	dbBind                        = wire.Bind(new(repository.DB), new(*gorm2.DB))
	launcherSessionRepositoryBind = wire.Bind(new(repository.LauncherSession), new(*gorm2.LauncherSession))
	launcherUserRepositoryBind    = wire.Bind(new(repository.LauncherUser), new(*gorm2.LauncherUser))
	launcherVersionRepositoryBind = wire.Bind(new(repository.LauncherVersion), new(*gorm2.LauncherVersion))

	launcherAuthServiceBind = wire.Bind(new(service.LauncherAuth), new(*v1Service.LauncherAuth))

	isProductionField  = wire.FieldsOf(new(*Config), "IsProduction")
	sessionKeyField    = wire.FieldsOf(new(*Config), "SessionKey")
	sessionSecretField = wire.FieldsOf(new(*Config), "SessionSecret")
)

func InjectAPI(config *Config) (*v1Handler.API, error) {
	wire.Build(
		isProductionField,
		sessionKeyField,
		sessionSecretField,
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
		v1Handler.NewSession,
		v1Handler.NewLauncherAuth,
		v1Handler.NewMiddleware,
	)
	return nil, nil
}
