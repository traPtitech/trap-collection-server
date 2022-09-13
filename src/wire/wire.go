//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/traPtitech/trap-collection-server/src/handler"
	"github.com/traPtitech/trap-collection-server/src/repository"
)

type App struct {
	*handler.API
	repository.DB
}

func newApp(api *handler.API, db repository.DB) *App {
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
