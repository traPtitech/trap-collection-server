package v1

type API struct {
	*Middleware
	*LauncherAuth
	*OAuth2
	*Session
}

func NewAPI(launcherAuth *LauncherAuth, oAuth2 *OAuth2, middleware *Middleware, session *Session) *API {
	return &API{
		LauncherAuth: launcherAuth,
		OAuth2:       oAuth2,
		Middleware:   middleware,
		Session:      session,
	}
}
