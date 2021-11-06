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
	"github.com/traPtitech/trap-collection-server/src/storage"
	"github.com/traPtitech/trap-collection-server/src/storage/local"
	"github.com/traPtitech/trap-collection-server/src/storage/swift"
)

type Config struct {
	IsProduction    common.IsProduction
	SessionKey      common.SessionKey
	SessionSecret   common.SessionSecret
	TraQBaseURL     common.TraQBaseURL
	OAuthClientID   common.ClientID
	Administrators  common.Administrators
	SwiftAuthURL    common.SwiftAuthURL
	SwiftUserName   common.SwiftUserName
	SwiftPassword   common.SwiftPassword
	SwiftTenantID   common.SwiftTenantID
	SwiftTenantName common.SwiftTenantName
	SwiftContainer  common.SwiftContainer
	FilePath        common.FilePath
	HttpClient      *http.Client
}

type Storage struct {
	GameImage storage.GameImage
	GameVideo storage.GameVideo
}

func newStorage(
	gameImage storage.GameImage,
	gameVideo storage.GameVideo,
) *Storage {
	return &Storage{
		GameImage: gameImage,
		GameVideo: gameVideo,
	}
}

var (
	isProductionField    = wire.FieldsOf(new(*Config), "IsProduction")
	sessionKeyField      = wire.FieldsOf(new(*Config), "SessionKey")
	sessionSecretField   = wire.FieldsOf(new(*Config), "SessionSecret")
	traQBaseURLField     = wire.FieldsOf(new(*Config), "TraQBaseURL")
	oAuthClientIDField   = wire.FieldsOf(new(*Config), "OAuthClientID")
	administratorsField  = wire.FieldsOf(new(*Config), "Administrators")
	swiftAuthURLField    = wire.FieldsOf(new(*Config), "SwiftAuthURL")
	swiftUserNameField   = wire.FieldsOf(new(*Config), "SwiftUserName")
	swiftPasswordField   = wire.FieldsOf(new(*Config), "SwiftPassword")
	swiftTenantIDField   = wire.FieldsOf(new(*Config), "SwiftTenantID")
	swiftTenantNameField = wire.FieldsOf(new(*Config), "SwiftTenantName")
	swiftContainerField  = wire.FieldsOf(new(*Config), "SwiftContainer")
	filePathField        = wire.FieldsOf(new(*Config), "FilePath")
	httpClientField      = wire.FieldsOf(new(*Config), "HttpClient")

	gameImageField = wire.FieldsOf(new(*Storage), "GameImage")
	gameVideoField = wire.FieldsOf(new(*Storage), "GameVideo")
)

func injectedStorage(config *Config) (*Storage, error) {
	if config.IsProduction {
		return injectSwiftStorage(config)
	}

	return injectLocalStorage(config)
}

func injectSwiftStorage(config *Config) (*Storage, error) {
	wire.Build(
		swiftAuthURLField,
		swiftUserNameField,
		swiftPasswordField,
		swiftTenantIDField,
		swiftTenantNameField,
		swiftContainerField,
		filePathField,
		wire.Bind(new(storage.GameImage), new(*swift.GameImage)),
		wire.Bind(new(storage.GameVideo), new(*swift.GameVideo)),
		swift.NewClient,
		swift.NewGameImage,
		swift.NewGameVideo,
		newStorage,
	)

	return nil, nil
}

func injectLocalStorage(config *Config) (*Storage, error) {
	wire.Build(
		filePathField,
		wire.Bind(new(storage.GameImage), new(*local.GameImage)),
		wire.Bind(new(storage.GameVideo), new(*local.GameVideo)),
		local.NewDirectoryManager,
		local.NewGameImage,
		local.NewGameVideo,
		newStorage,
	)

	return nil, nil
}

var (
	dbBind                        = wire.Bind(new(repository.DB), new(*gorm2.DB))
	gameRepositoryBind            = wire.Bind(new(repository.Game), new(*gorm2.Game))
	gameImageRepositoryBind       = wire.Bind(new(repository.GameImage), new(*gorm2.GameImage))
	gameVideoRepositoryBind       = wire.Bind(new(repository.GameVideo), new(*gorm2.GameVideo))
	gameManagementRoleBind        = wire.Bind(new(repository.GameManagementRole), new(*gorm2.GameManagementRole))
	launcherSessionRepositoryBind = wire.Bind(new(repository.LauncherSession), new(*gorm2.LauncherSession))
	launcherUserRepositoryBind    = wire.Bind(new(repository.LauncherUser), new(*gorm2.LauncherUser))
	launcherVersionRepositoryBind = wire.Bind(new(repository.LauncherVersion), new(*gorm2.LauncherVersion))

	oidcAuthBind = wire.Bind(new(auth.OIDC), new(*traq.OIDC))
	userAuthBind = wire.Bind(new(auth.User), new(*traq.User))

	userCacheBind = wire.Bind(new(cache.User), new(*ristretto.User))

	administratorAuthServiceBind = wire.Bind(new(service.AdministratorAuth), new(*v1Service.AdministratorAuth))
	gameAuthServiceBind          = wire.Bind(new(service.GameAuth), new(*v1Service.GameAuth))
	gameImageServiceBind         = wire.Bind(new(service.GameImage), new(*v1Service.GameImage))
	gameVideoServiceBind         = wire.Bind(new(service.GameVideo), new(*v1Service.GameVideo))
	launcherAuthServiceBind      = wire.Bind(new(service.LauncherAuth), new(*v1Service.LauncherAuth))
	oidcServiceBind              = wire.Bind(new(service.OIDC), new(*v1Service.OIDC))
	userServiceBind              = wire.Bind(new(service.User), new(*v1Service.User))
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
		gameImageField,
		gameVideoField,
		dbBind,
		gameRepositoryBind,
		gameImageRepositoryBind,
		gameVideoRepositoryBind,
		gameManagementRoleBind,
		launcherSessionRepositoryBind,
		launcherUserRepositoryBind,
		launcherVersionRepositoryBind,
		oidcAuthBind,
		userAuthBind,
		userCacheBind,
		administratorAuthServiceBind,
		gameAuthServiceBind,
		gameImageServiceBind,
		gameVideoServiceBind,
		launcherAuthServiceBind,
		oidcServiceBind,
		userServiceBind,
		gorm2.NewDB,
		gorm2.NewGame,
		gorm2.NewGameImage,
		gorm2.NewGameVideo,
		gorm2.NewGameManagementRole,
		gorm2.NewLauncherSession,
		gorm2.NewLauncherUser,
		gorm2.NewLauncherVersion,
		traq.NewOIDC,
		traq.NewUser,
		ristretto.NewUser,
		v1Service.NewAdministratorAuth,
		v1Service.NewGameAuth,
		v1Service.NewGameImage,
		v1Service.NewGameVideo,
		v1Service.NewLauncherAuth,
		v1Service.NewOIDC,
		v1Service.NewUser,
		v1Service.NewUserUtils,
		v1Handler.NewAPI,
		v1Handler.NewSession,
		v1Handler.NewGameRole,
		v1Handler.NewGameImage,
		v1Handler.NewGameVideo,
		v1Handler.NewLauncherAuth,
		v1Handler.NewOAuth2,
		v1Handler.NewUser,
		v1Handler.NewMiddleware,
		injectedStorage,
	)
	return nil, nil
}
