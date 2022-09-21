package v2

//go:generate sh -c "go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config ./openapi/config.yaml ../../../docs/openapi/v2.yaml > openapi/openapi.gen.go"
//go:generate go fmt ./openapi/openapi.gen.go

import (
	"fmt"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

type API struct {
	*Checker
	*Session
	*OAuth2
	*User
	*Admin
	*Game
	*GameRole
	*GameVersion
	*GameFile
	*GameImage
	*GameVideo
	*Edition
	*EditionAuth
	*Seat
}

func NewAPI(
	checker *Checker,
	session *Session,
	oAuth2 *OAuth2,
	user *User,
	admin *Admin,
	game *Game,
	gameRole *GameRole,
	gameVersion *GameVersion,
	gameFile *GameFile,
	gameImage *GameImage,
	gameVideo *GameVideo,
	edition *Edition,
	editionAuth *EditionAuth,
	seat *Seat,
) *API {
	return &API{
		Checker:     checker,
		Session:     session,
		OAuth2:      oAuth2,
		User:        user,
		Admin:       admin,
		Game:        game,
		GameRole:    gameRole,
		GameVersion: gameVersion,
		GameFile:    gameFile,
		GameImage:   gameImage,
		GameVideo:   gameVideo,
		Edition:     edition,
		EditionAuth: editionAuth,
		Seat:        seat,
	}
}

func (api *API) SetRoutes(e *echo.Echo) error {
	return api.setRoutes(e)
}

func (api *API) Start(addr string) error {
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

	return e.Start(addr)
}

func (api *API) setRoutes(e *echo.Echo) error {
	swagger, err := openapi.GetSwagger()
	if err != nil {
		return fmt.Errorf("failed to get openapi: %w", err)
	}
	apiGroup := e.Group("/api/v2")
	apiGroup.Use(oapiMiddleware.OapiRequestValidatorWithOptions(swagger, &oapiMiddleware.Options{
		Options: openapi3filter.Options{
			MultiError:         true,
			AuthenticationFunc: api.Checker.check,
		},
	}))
	openapi.RegisterHandlers(apiGroup, api)

	return nil
}
