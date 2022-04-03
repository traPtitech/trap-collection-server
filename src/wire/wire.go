//go:build wireinject

package wire

import (
	"github.com/google/wire"
	v1Handler "github.com/traPtitech/trap-collection-server/src/handler/v1"
	"github.com/traPtitech/trap-collection-server/src/repository"
)

type App struct {
	*v1Handler.API
	repository.DB
}

func newApp(api *v1Handler.API, db repository.DB) *App {
	return &App{
		API: api,
		DB:  db,
	}
}

func (app *App) Run() error {
	defer app.DB.Close()

	return app.API.Start()
}

func InjectApp() (*App, error) {
	wire.Build(
		configSet,
		serviceSet,
		authSet,
		cacheSet,
		handlerSet,
		repositorySet,
		storageSet,

		newApp,
	)

	return nil, nil
}
