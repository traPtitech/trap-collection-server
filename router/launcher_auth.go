package router

import "github.com/traPtitech/trap-collection-server/openapi"

type LauncherAuth struct {
	openapi.LauncherAuthApi
}

func newLauncherAuth() *LauncherAuth {
	return &LauncherAuth{}
}
