package v1

type API struct {
	*LauncherAuth
}

func NewAPI(launcherAuth *LauncherAuth) *API {
	return &API{
		LauncherAuth: launcherAuth,
	}
}
