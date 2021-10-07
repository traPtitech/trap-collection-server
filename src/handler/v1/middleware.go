package v1

import "github.com/traPtitech/trap-collection-server/src/service"

type Middleware struct {
	launcherAuthService service.LauncherAuth
}

func NewMiddleware(launcherAuthService service.LauncherAuth) *Middleware {
	return &Middleware{
		launcherAuthService: launcherAuthService,
	}
}
