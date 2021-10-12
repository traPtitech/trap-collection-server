package v1

type API struct {
	*Middleware
	*User
	*LauncherAuth
	*OAuth2
	*Session
}

func NewAPI(user *User, launcherAuth *LauncherAuth, oAuth2 *OAuth2, middleware *Middleware, session *Session) *API {
	return &API{
		User:         user,
		LauncherAuth: launcherAuth,
		OAuth2:       oAuth2,
		Middleware:   middleware,
		Session:      session,
	}
}
