package v1

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
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

func (ga *GameAuth) AddGameCollaborators(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, userIDs []values.TraPMemberID) error {
	err := ga.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := ga.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		users, err := ga.userUtils.getAllActiveUser(ctx, session)
		if err != nil {
			return fmt.Errorf("failed to get active users: %v", err)
		}

		userMap := make(map[values.TraPMemberID]struct{}, len(users))
		for _, user := range users {
			userMap[user.GetID()] = struct{}{}
		}

		invalidUserIDs := []string{}
		for _, userID := range userIDs {
			if _, ok := userMap[userID]; !ok {
				invalidUserIDs = append(invalidUserIDs, uuid.UUID(userID).String())
			}
		}

		if len(invalidUserIDs) != 0 {
			return fmt.Errorf("invalid userID(%s): %w", strings.Join(invalidUserIDs, ", "), service.ErrInvalidUserID)
		}

		err = ga.gameManagementRoleRepository.AddGameManagementRoles(
			ctx,
			gameID,
			userIDs,
			values.GameManagementRoleCollaborator,
		)
		if err != nil {
			return fmt.Errorf("failed to add game management roles: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}

func (ga *GameAuth) UpdateGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID, role values.GameManagementRole) error {
	err := ga.db.Transaction(ctx, nil, func(ctx context.Context) error {
		nowRole, err := ga.gameManagementRoleRepository.GetGameManagementRole(
			ctx,
			gameID,
			userID,
			repository.LockTypeRecord,
		)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidRole
		}
		if err != nil {
			return fmt.Errorf("failed to get game management role: %w", err)
		}

		if role == nowRole {
			return service.ErrNoGameManagementRoleUpdated
		}

		err = ga.gameManagementRoleRepository.UpdateGameManagementRole(ctx, gameID, userID, role)
		if err != nil {
			return fmt.Errorf("failed to update game management role: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed in transaction: %w", err)
	}

	return nil
}

func (ga *GameAuth) RemoveGameCollaborator(ctx context.Context, gameID values.GameID, userID values.TraPMemberID) error {
	err := ga.db.Transaction(ctx, nil, func(ctx context.Context) error {
		role, err := ga.gameManagementRoleRepository.GetGameManagementRole(
			ctx,
			gameID,
			userID,
			repository.LockTypeRecord,
		)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidRole
		}
		if err != nil {
			return fmt.Errorf("failed to get game management role: %w", err)
		}

		if role != values.GameManagementRoleCollaborator {
			return service.ErrInvalidRole
		}

		err = ga.gameManagementRoleRepository.RemoveGameManagementRole(ctx, gameID, userID)
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

func (ga *GameAuth) GetGameManagers(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) ([]*service.GameManager, error) {
	_, err := ga.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	userIDAndRoles, err := ga.gameManagementRoleRepository.GetGameManagersByGameID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game managers: %w", err)
	}

	users, err := ga.userUtils.getAllActiveUser(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %v", err)
	}

	userMap := make(map[values.TraPMemberID]*service.UserInfo, len(users))
	for _, user := range users {
		userMap[user.GetID()] = user
	}

	gameManagers := make([]*service.GameManager, 0, len(userIDAndRoles))
	for _, userIDAndRole := range userIDAndRoles {
		user, ok := userMap[userIDAndRole.UserID]
		if !ok {
			// 凍結された管理者がいる可能性があるので、エラーにはしない
			continue
		}

		gameManagers = append(gameManagers, &service.GameManager{
			UserID:     user.GetID(),
			UserName:   user.GetName(),
			UserStatus: user.GetStatus(),
			Role:       userIDAndRole.Role,
		})
	}

	return gameManagers, nil
}
