package v1

type API struct {
	*Middleware
	*LauncherAuth
	*Session
}

func NewAPI(launcherAuth *LauncherAuth, middleware *Middleware, session *Session) *API {
	return &API{
		LauncherAuth: launcherAuth,
		Middleware:   middleware,
		Session:      session,
	}
}
