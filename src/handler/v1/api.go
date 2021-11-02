package v1

type API struct {
	*Middleware
	*User
	*GameRole
	*LauncherAuth
	*OAuth2
	*Session
}

func NewAPI(
	middleware *Middleware,
	user *User,
	gameRole *GameRole,
	launcherAuth *LauncherAuth,
	oAuth2 *OAuth2,
	session *Session,
) *API {
	return &API{
		Middleware:   middleware,
		User:         user,
		GameRole:     gameRole,
		LauncherAuth: launcherAuth,
		OAuth2:       oAuth2,
		Session:      session,
	}
}
