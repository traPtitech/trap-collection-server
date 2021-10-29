package v1

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
)

type GameAuth struct {
	db                           repository.DB
	gameRepository               repository.Game
	gameManagementRoleRepository repository.GameManagementRole
	userUtils                    *UserUtils
}

func NewGameAuth(
	db repository.DB,
	gameRepository repository.Game,
	gameManagementRoleRepository repository.GameManagementRole,
	userUtils *UserUtils,
) *GameAuth {
	return &GameAuth{
		db:                           db,
		gameRepository:               gameRepository,
		gameManagementRoleRepository: gameManagementRoleRepository,
		userUtils:                    userUtils,
	}
}
