package v1

import (
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type LauncherAuth struct {
	launcherAuthService service.LauncherAuth
	openapi.LauncherAuthApi
}

func NewLauncherAuth(launcherAuthService service.LauncherAuth) *LauncherAuth {
	return &LauncherAuth{
		launcherAuthService: launcherAuthService,
	}
}
