package v1

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
)

type LauncherAuth struct {
	db                        repository.DB
	launcherVersionRepository repository.LauncherVersion
	launcherUserRepository    repository.LauncherUser
	launcherSessionRepository repository.LauncherSession
}

func NewLauncherAuth(
	db repository.DB,
	launcherVersionRepository repository.LauncherVersion,
	launcherUserRepository repository.LauncherUser,
	launcherSessionRepository repository.LauncherSession,
) *LauncherAuth {
	return &LauncherAuth{
		db:                        db,
		launcherVersionRepository: launcherVersionRepository,
		launcherUserRepository:    launcherUserRepository,
		launcherSessionRepository: launcherSessionRepository,
	}
}
