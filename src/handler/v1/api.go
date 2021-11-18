package v1

type API struct {
	*Middleware
	*User
	*GameRole
	*GameImage
	*GameVideo
	*GameVersion
	*LauncherAuth
	*OAuth2
	*Session
}

func NewAPI(
	middleware *Middleware,
	user *User,
	gameRole *GameRole,
	gameImage *GameImage,
	gameVideo *GameVideo,
	gameVersion *GameVersion,
	launcherAuth *LauncherAuth,
	oAuth2 *OAuth2,
	session *Session,
) *API {
	return &API{
		Middleware:   middleware,
		User:         user,
		GameRole:     gameRole,
		GameImage:    gameImage,
		GameVideo:    gameVideo,
		GameVersion:  gameVersion,
		LauncherAuth: launcherAuth,
		OAuth2:       oAuth2,
		Session:      session,
	}
}
