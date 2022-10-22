package v2

import (
	"github.com/traPtitech/trap-collection-server/src/repository"
)

type GameRole struct {
	db                           repository.DB
	gameRepository               repository.GameV2
	gameManagementRoleRepository repository.GameManagementRole
	userUtils                    *User
}

func NewGameRole(
	db repository.DB,
	gameRepository repository.GameV2,
	gameManagementRoleRepository repository.GameManagementRole,
	userUtils *User,
) *GameRole {
	return &GameRole{
		db:                           db,
		gameRepository:               gameRepository,
		gameManagementRoleRepository: gameManagementRoleRepository,
		userUtils:                    userUtils,
	}
}
