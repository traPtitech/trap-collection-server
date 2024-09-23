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
	user                         *User
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
		user:                         userUtils,
	}
}

func (gameRole *GameRole) EditGameManagementRole(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, userID values.TraPMemberID, newRole values.GameManagementRole) error {
	err := gameRole.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gameRole.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGame
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		activeUsers, err := gameRole.user.getActiveUsers(ctx, session)
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
		ownersNumber := 0
		for _, managerAndRole := range gameManagers {
			gameManagersMap[managerAndRole.UserID] = managerAndRole.Role
			if managerAndRole.Role == values.GameManagementRoleAdministrator {
				ownersNumber++
			}
		}

		if role, ok := gameManagersMap[userID]; ok {
			if role != newRole { //既にあるroleと違うので、Update
				if role == values.GameManagementRoleAdministrator && ownersNumber == 1 { //ownersが一人の場合にそのownerをmaintainerに変えるのを止める。
					return service.ErrCannotEditOwners
				}
				err = gameRole.gameManagementRoleRepository.UpdateGameManagementRole(ctx, gameID, userID, newRole)
				if errors.Is(err, repository.ErrNoRecordUpdated) {
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

func (gameRole *GameRole) RemoveGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID) error {
	err := gameRole.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gameRole.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGame
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}
		managers, err := gameRole.gameManagementRoleRepository.GetGameManagersByGameID(
			ctx,
			gameID,
		)
		if err != nil {
			return fmt.Errorf("failed to get game managers by gameID: %w", err)
		}

		managersMap := make(map[values.TraPMemberID]values.GameManagementRole, len(managers))
		ownersNumber := 0
		for _, manager := range managers {
			managersMap[manager.UserID] = manager.Role
			if manager.Role == values.GameManagementRoleAdministrator {
				ownersNumber++
			}
		}

		if _, ok := managersMap[userID]; !ok {
			return service.ErrInvalidRole
		}
		if managersMap[userID] == values.GameManagementRoleAdministrator && ownersNumber == 1 {
			return service.ErrCannotDeleteOwner
		}

		err = gameRole.gameManagementRoleRepository.RemoveGameManagementRole(ctx, gameID, userID)
		if err != nil {
			return fmt.Errorf("failed to remove game management role: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}

func (gameRole *GameRole) UpdateGameAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error {
	myInfo, err := gameRole.user.getMe(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to get me: %w", err)
	}

	_, err = gameRole.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return service.ErrNoGame
	}
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	role, err := gameRole.gameManagementRoleRepository.GetGameManagementRole(ctx, gameID, myInfo.GetID(), repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return service.ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("failed to get game management role: %w", err)
	}

	if !role.HaveGameUpdatePermission() {
		return service.ErrForbidden
	}

	return nil
}

func (gameRole *GameRole) UpdateGameManagementRoleAuth(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) error {
	myInfo, err := gameRole.user.getMe(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to get me: %w", err)
	}

	_, err = gameRole.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return service.ErrNoGame
	}
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	role, err := gameRole.gameManagementRoleRepository.GetGameManagementRole(ctx, gameID, myInfo.GetID(), repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return service.ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("failed to get game management role: %w", err)
	}

	if !role.HaveUpdateManagementRolePermission() {
		return service.ErrForbidden
	}

	return nil
}
