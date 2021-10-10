//go:build wireinject
// +build wireinject

package src

import (
	"net/http"

	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/pkg/common"
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/auth/traQ"
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
	TraQBaseURL   common.TraQBaseURL
	OAuthClientID common.ClientID
	HttpClient    *http.Client
}

var (
	dbBind                        = wire.Bind(new(repository.DB), new(*gorm2.DB))
	launcherSessionRepositoryBind = wire.Bind(new(repository.LauncherSession), new(*gorm2.LauncherSession))
	launcherUserRepositoryBind    = wire.Bind(new(repository.LauncherUser), new(*gorm2.LauncherUser))
	launcherVersionRepositoryBind = wire.Bind(new(repository.LauncherVersion), new(*gorm2.LauncherVersion))

	oidcAuthBind = wire.Bind(new(auth.OIDC), new(*traq.OIDC))

	launcherAuthServiceBind = wire.Bind(new(service.LauncherAuth), new(*v1Service.LauncherAuth))
	oidcServiceBind         = wire.Bind(new(service.OIDC), new(*v1Service.OIDC))

	isProductionField  = wire.FieldsOf(new(*Config), "IsProduction")
	sessionKeyField    = wire.FieldsOf(new(*Config), "SessionKey")
	sessionSecretField = wire.FieldsOf(new(*Config), "SessionSecret")
	traQBaseURLField   = wire.FieldsOf(new(*Config), "TraQBaseURL")
	oAuthClientIDField = wire.FieldsOf(new(*Config), "OAuthClientID")
	httpClientField    = wire.FieldsOf(new(*Config), "HttpClient")
)

func InjectAPI(config *Config) (*v1Handler.API, error) {
	wire.Build(
		isProductionField,
		sessionKeyField,
		sessionSecretField,
		traQBaseURLField,
		oAuthClientIDField,
		httpClientField,
		dbBind,
		launcherSessionRepositoryBind,
		launcherUserRepositoryBind,
		launcherVersionRepositoryBind,
		oidcAuthBind,
		launcherAuthServiceBind,
		oidcServiceBind,
		gorm2.NewDB,
		gorm2.NewLauncherSession,
		gorm2.NewLauncherUser,
		gorm2.NewLauncherVersion,
		traq.NewOIDC,
		v1Service.NewLauncherAuth,
		v1Service.NewOIDC,
		v1Handler.NewAPI,
		v1Handler.NewSession,
		v1Handler.NewLauncherAuth,
		v1Handler.NewOAuth2,
		v1Handler.NewMiddleware,
	)
	return nil, nil
}
