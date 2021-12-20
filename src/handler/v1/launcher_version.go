package v1

import (
	"github.com/traPtitech/trap-collection-server/src/service"
)

type LauncherVersion struct {
	launcherVersionService service.LauncherVersion
}

func NewLauncherVersion(launcherVersionService service.LauncherVersion) *LauncherVersion {
	return &LauncherVersion{
		launcherVersionService: launcherVersionService,
	}
}
