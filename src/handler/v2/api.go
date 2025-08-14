package v2

//go:generate sh -c "go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config ./openapi/config.yaml ../../../docs/openapi/v2.yaml > openapi/openapi.gen.go"
//go:generate go fmt ./openapi/openapi.gen.go

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
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
	*GameGenre
	*GameVersion
	*GameFile
	*GameImage
	*GameVideo
	*GamePlayLog
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
	gameGenre *GameGenre,
	gameVersion *GameVersion,
	gameFile *GameFile,
	gameImage *GameImage,
	gameVideo *GameVideo,
	gamePlayLog *GamePlayLog,
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
		GameGenre:   gameGenre,
		GameVersion: gameVersion,
		GameFile:    gameFile,
		GameImage:   gameImage,
		GameVideo:   gameVideo,
		GamePlayLog: gamePlayLog,
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

	// oapiMiddleware.OapiRequestValidatorの設定されるパスが"/"でないと正常に動作しないが、
	// 他のrouteにはoapiMiddleware.OapiRequestValidatorを設定したくないため、
	// 空のpathのgroupを作成し、oapiMiddleware.OapiRequestValidatorを設定する
	apiGroup := e.Group("")
	apiGroup.Use(echomiddleware.OapiRequestValidatorWithOptions(swagger, &echomiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: api.Checker.check,
			// validate時にデータがメモリに乗るため、
			// 画像・動画・ファイルのような大きなデータのアップロード時にメモリ不足にならないように、
			// ExcludeRequestBody、ExcludeResponseBodyをtrueにする
			ExcludeRequestBody:  true,
			ExcludeResponseBody: true,
		},
	}))
	openapi.RegisterHandlersWithBaseURL(apiGroup, api, "/api/v2")

	return nil
}
