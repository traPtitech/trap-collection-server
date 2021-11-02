//go:build wireinject
// +build wireinject

package src

import (
	"net/http"

	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/auth/traQ"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/cache/ristretto"
	v1Handler "github.com/traPtitech/trap-collection-server/src/handler/v1"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2"
	"github.com/traPtitech/trap-collection-server/src/service"
	v1Service "github.com/traPtitech/trap-collection-server/src/service/v1"
)

type Config struct {
	IsProduction   common.IsProduction
	SessionKey     common.SessionKey
	SessionSecret  common.SessionSecret
	TraQBaseURL    common.TraQBaseURL
	OAuthClientID  common.ClientID
	Administrators common.Administrators
	HttpClient     *http.Client
}

var (
	dbBind                        = wire.Bind(new(repository.DB), new(*gorm2.DB))
	gameRepositoryBind            = wire.Bind(new(repository.Game), new(*gorm2.Game))
	gameManagementRoleBind        = wire.Bind(new(repository.GameManagementRole), new(*gorm2.GameManagementRole))
	launcherSessionRepositoryBind = wire.Bind(new(repository.LauncherSession), new(*gorm2.LauncherSession))
	launcherUserRepositoryBind    = wire.Bind(new(repository.LauncherUser), new(*gorm2.LauncherUser))
	launcherVersionRepositoryBind = wire.Bind(new(repository.LauncherVersion), new(*gorm2.LauncherVersion))

	oidcAuthBind = wire.Bind(new(auth.OIDC), new(*traq.OIDC))
	userAuthBind = wire.Bind(new(auth.User), new(*traq.User))

	userCacheBind = wire.Bind(new(cache.User), new(*ristretto.User))

	administratorAuthServiceBind = wire.Bind(new(service.AdministratorAuth), new(*v1Service.AdministratorAuth))
	gameAuthServiceBind          = wire.Bind(new(service.GameAuth), new(*v1Service.GameAuth))
	launcherAuthServiceBind      = wire.Bind(new(service.LauncherAuth), new(*v1Service.LauncherAuth))
	oidcServiceBind              = wire.Bind(new(service.OIDC), new(*v1Service.OIDC))
	userServiceBind              = wire.Bind(new(service.User), new(*v1Service.User))

	isProductionField   = wire.FieldsOf(new(*Config), "IsProduction")
	sessionKeyField     = wire.FieldsOf(new(*Config), "SessionKey")
	sessionSecretField  = wire.FieldsOf(new(*Config), "SessionSecret")
	traQBaseURLField    = wire.FieldsOf(new(*Config), "TraQBaseURL")
	oAuthClientIDField  = wire.FieldsOf(new(*Config), "OAuthClientID")
	administratorsField = wire.FieldsOf(new(*Config), "Administrators")
	httpClientField     = wire.FieldsOf(new(*Config), "HttpClient")
)

func InjectAPI(config *Config) (*v1Handler.API, error) {
	wire.Build(
		isProductionField,
		sessionKeyField,
		sessionSecretField,
		traQBaseURLField,
		oAuthClientIDField,
		administratorsField,
		httpClientField,
		dbBind,
		gameRepositoryBind,
		gameManagementRoleBind,
		launcherSessionRepositoryBind,
		launcherUserRepositoryBind,
		launcherVersionRepositoryBind,
		oidcAuthBind,
		userAuthBind,
		userCacheBind,
		administratorAuthServiceBind,
		gameAuthServiceBind,
		launcherAuthServiceBind,
		oidcServiceBind,
		userServiceBind,
		gorm2.NewDB,
		gorm2.NewGame,
		gorm2.NewGameManagementRole,
		gorm2.NewLauncherSession,
		gorm2.NewLauncherUser,
		gorm2.NewLauncherVersion,
		traq.NewOIDC,
		traq.NewUser,
		ristretto.NewUser,
		v1Service.NewAdministratorAuth,
		v1Service.NewGameAuth,
		v1Service.NewLauncherAuth,
		v1Service.NewOIDC,
		v1Service.NewUser,
		v1Service.NewUserUtils,
		v1Handler.NewAPI,
		v1Handler.NewSession,
		v1Handler.NewGameRole,
		v1Handler.NewLauncherAuth,
		v1Handler.NewOAuth2,
		v1Handler.NewUser,
		v1Handler.NewMiddleware,
	)
	return nil, nil
}
