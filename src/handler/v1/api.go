package v1

type API struct {
	*Middleware
	*LauncherAuth
}

func NewAPI(launcherAuth *LauncherAuth, middleware *Middleware) *API {
	return &API{
		LauncherAuth: launcherAuth,
		Middleware:   middleware,
	}
}
