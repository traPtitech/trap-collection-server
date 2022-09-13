package handler

import (
	"fmt"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/handler/common"
	v1 "github.com/traPtitech/trap-collection-server/src/handler/v1"
	v2 "github.com/traPtitech/trap-collection-server/src/handler/v2"
)

type API struct {
	featureV2 bool
	addr      string
	session   *common.Session
	v1        *v1.API
	v2        *v2.API
}

func NewAPI(appConf config.App, conf config.Handler, session *common.Session, v1 *v1.API, v2 *v2.API) (*API, error) {
	addr, err := conf.Addr()
	if err != nil {
		return nil, fmt.Errorf("failed to get addr: %w", err)
	}

	return &API{
		featureV2: appConf.FeatureV2(),
		addr:      addr,
		session:   session,
		v1:        v1,
		v2:        v2,
	}, nil
}

func (api *API) Start() error {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	p := prometheus.NewPrometheus("echo", nil)
	p.MetricsPath = "/api/metrics"
	p.Use(e)

	api.session.Use(e)

	err := api.v1.SetRoutes(e)
	if err != nil {
		return fmt.Errorf("failed to set v1 routes: %w", err)
	}

	if api.featureV2 {
		err := api.v2.SetRoutes(e)
		if err != nil {
			return fmt.Errorf("failed to set v2 routes: %w", err)
		}
	}

	return e.Start(api.addr)
}
