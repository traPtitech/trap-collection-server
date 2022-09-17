package v1

import (
	"errors"
	"log"
	"os"
	"strconv"

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

func (*App) FeatureV2() bool {
	env, ok := os.LookupEnv(envKeyFeatureV2)
	if !ok {
		return false
	}

	v2, err := strconv.ParseBool(env)
	if err != nil {
		log.Printf("failed to parse %s: %v\n", envKeyFeatureV2, err)
		return false
	}

	return v2
}
