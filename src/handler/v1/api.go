package v1

import (
	"fmt"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
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
	addr string
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
	conf config.Handler,
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
) (*API, error) {
	addr, err := conf.Addr()
	if err != nil {
		return nil, fmt.Errorf("failed to get addr: %w", err)
	}

	return &API{
		addr:            addr,
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
	}, nil
}

func (api *API) SetRoutes(e *echo.Echo) error {
	return api.setRoutes(e)
}

func (api *API) Start() error {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	p := prometheus.NewPrometheus("echo", nil)
	p.MetricsPath = "/api/metrics"
	p.Use(e)

	api.Session.Use(e)
	err := api.setRoutes(e)
	if err != nil {
		return fmt.Errorf("failed to set routes: %w", err)
	}

	return e.Start(api.addr)
}

func (api *API) setRoutes(e *echo.Echo) error {
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

	openapi.SetupRouting(e, openapiAPI)

	return nil
}
