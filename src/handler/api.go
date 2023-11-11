package handler

import (
	"fmt"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/handler/common"

	v2 "github.com/traPtitech/trap-collection-server/src/handler/v2"
)

type API struct {
	addr    string
	session *common.Session
	v2      *v2.API
}

func NewAPI(appConf config.App, conf config.Handler, session *common.Session, v2 *v2.API) (*API, error) {
	addr, err := conf.Addr()
	if err != nil {
		return nil, fmt.Errorf("failed to get addr: %w", err)
	}

	if !appConf.FeatureV2() {
		return nil, fmt.Errorf("only v2 is allowed")
	}

	return &API{
		addr:    addr,
		session: session,
		v2:      v2,
	}, nil
}

func (api *API) Start() error {
	const metricsPath = "/api/metrics"

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == metricsPath
		},
	}))

	p := prometheus.NewPrometheus("echo", nil)
	p.MetricsPath = metricsPath
	p.Use(e)

	api.session.Use(e)

	err := api.v2.SetRoutes(e)
	if err != nil {
		return fmt.Errorf("failed to set v2 routes: %w", err)
	}

	return e.Start(api.addr)
}
