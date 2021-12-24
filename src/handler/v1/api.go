package v1

import (
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/traPtitech/trap-collection-server/openapi"
)

type GameAPI struct {
	*Game
	*GameRole
	*GameImage
	*GameVideo
	*GameVersion
	*GameFile
	*GameURL
}

type API struct {
	*Middleware
	*User
	*Game
	*GameRole
	*GameImage
	*GameVideo
	*GameVersion
	*GameFile
	*GameURL
	*LauncherAuth
	*LauncherVersion
	*OAuth2
	*Session
}

func NewAPI(
	middleware *Middleware,
	user *User,
	game *Game,
	gameRole *GameRole,
	gameImage *GameImage,
	gameVideo *GameVideo,
	gameVersion *GameVersion,
	gameFile *GameFile,
	gameURL *GameURL,
	launcherAuth *LauncherAuth,
	launcherVersion *LauncherVersion,
	oAuth2 *OAuth2,
	session *Session,
) *API {
	return &API{
		Middleware:      middleware,
		User:            user,
		Game:            game,
		GameRole:        gameRole,
		GameImage:       gameImage,
		GameVideo:       gameVideo,
		GameVersion:     gameVersion,
		GameFile:        gameFile,
		GameURL:         gameURL,
		LauncherAuth:    launcherAuth,
		LauncherVersion: launcherVersion,
		OAuth2:          oAuth2,
		Session:         session,
	}
}

func (api *API) Start(addr string) error {
	openapiAPI := &openapi.Api{
		Middleware: api.Middleware,
		GameApi: GameAPI{
			Game:        api.Game,
			GameRole:    api.GameRole,
			GameImage:   api.GameImage,
			GameVideo:   api.GameVideo,
			GameVersion: api.GameVersion,
			GameFile:    api.GameFile,
			GameURL:     api.GameURL,
		},
		LauncherAuthApi: api.LauncherAuth,
		Oauth2Api:       api.OAuth2,
		UserApi:         api.User,
		VersionApi:      api.LauncherVersion,
	}

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	p := prometheus.NewPrometheus("echo", nil)
	p.MetricsPath = "/api/metrics"
	p.Use(e)

	api.Session.Use(e)

	openapi.SetupRouting(e, openapiAPI)

	return e.Start(addr)
}
