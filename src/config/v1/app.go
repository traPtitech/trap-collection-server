package v1

import (
	"errors"
	"os"

	"github.com/traPtitech/trap-collection-server/src/config"
)

type App struct{}

func NewApp() *App {
	return &App{}
}

func (*App) Status() (config.AppStatus, error) {
	env, ok := os.LookupEnv(envKeyCollectionEnv)
	if !ok {
		return config.AppStatusProduction, nil
	}

	switch env {
	case "production":
		return config.AppStatusProduction, nil
	case "development":
		return config.AppStatusDevelopment, nil
	}

	return 0, errors.New("invalid env")
}
