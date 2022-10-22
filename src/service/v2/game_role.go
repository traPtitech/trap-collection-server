package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
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

func (gameRole *GameRole) EditGameManagementRole(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, userID values.TraPMemberID, newRole values.GameManagementRole) error {
	err := gameRole.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gameRole.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		activeUsers, err := gameRole.userUtils.getActiveUsers(ctx, session)
		if err != nil {
			return fmt.Errorf("failed to get active users: %v", err)
		}
		activeUsersMap := make(map[values.TraPMemberID]struct{}, len(activeUsers))
		for _, activeUser := range activeUsers {
			activeUsersMap[activeUser.GetID()] = struct{}{}
		}
		if _, ok := activeUsersMap[userID]; !ok {
			return service.ErrInvalidUserID
		}

		//ゲームの管理者をいったん全部取得
		gameManagers, err := gameRole.gameManagementRoleRepository.GetGameManagersByGameID(ctx, gameID)
		if err != nil {
			return fmt.Errorf("error: failed to get game managers by gameID: %w", err)
		}
		gameManagersMap := make(map[values.TraPMemberID]values.GameManagementRole, len(gameManagers))
		for _, managerAndRole := range gameManagers {
			gameManagersMap[managerAndRole.UserID] = managerAndRole.Role
		}

		if role, ok := gameManagersMap[userID]; ok {
			if role != newRole { //既にあるroleと違うので、Update
				err = gameRole.gameManagementRoleRepository.UpdateGameManagementRole(ctx, gameID, userID, newRole)
				if errors.Is(repository.ErrNoRecordUpdated, err) {
					return service.ErrNoGameManagementRoleUpdated
				}
				if err != nil {
					return fmt.Errorf("error: failed to update game management role: %w", err)
				}
			} else { //既にあるroleと同じなので、エラー
				return service.ErrNoGameManagementRoleUpdated
			}
		} else { //roleを持っていなかったので、追加する。
			err = gameRole.gameManagementRoleRepository.AddGameManagementRoles(ctx, gameID, []values.TraPMemberID{userID}, newRole)
			if err != nil {
				return fmt.Errorf("error: failed to add game management role: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}
	return nil
}
