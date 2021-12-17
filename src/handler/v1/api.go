package v1

type API struct {
	*Middleware
	*User
	*GameRole
	*GameImage
	*GameVideo
	*GameVersion
	*GameFile
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
	gameFile *GameFile,
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
		GameFile:     gameFile,
		LauncherAuth: launcherAuth,
		OAuth2:       oAuth2,
		Session:      session,
	}
}
