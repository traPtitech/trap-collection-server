package v1

type API struct {
	*Middleware
	*User
	*GameRole
	*GameImage
	*LauncherAuth
	*OAuth2
	*Session
}

func NewAPI(
	middleware *Middleware,
	user *User,
	gameRole *GameRole,
	gameImage *GameImage,
	launcherAuth *LauncherAuth,
	oAuth2 *OAuth2,
	session *Session,
) *API {
	return &API{
		Middleware:   middleware,
		User:         user,
		GameRole:     gameRole,
		GameImage:    gameImage,
		LauncherAuth: launcherAuth,
		OAuth2:       oAuth2,
		Session:      session,
	}
}
