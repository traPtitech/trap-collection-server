package v1

import "github.com/traPtitech/trap-collection-server/src/repository"

type LauncherVersion struct {
	db                        repository.DB
	launcherVersionRepository repository.LauncherVersion
	gameRepository            repository.Game
}

func NewLauncherVersion(
	db repository.DB,
	launcherVersionRepository repository.LauncherVersion,
	gameRepository repository.Game,
) *LauncherVersion {
	return &LauncherVersion{
		db:                        db,
		launcherVersionRepository: launcherVersionRepository,
		gameRepository:            gameRepository,
	}
}
