package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

const (
	gameManagementRoleTypeAdministrator = "administrator"
	gameManagementRoleTypeCollaborator  = "collaborator"
)

type GameManagementRole struct {
	db *DB
}

func NewGameManagementRole(db *DB) *GameManagementRole {
	return &GameManagementRole{
		db: db,
	}
}

func (gmr *GameManagementRole) AddGameManagementRoles(ctx context.Context, gameID values.GameID, userIDs []values.TraPMemberID, role values.GameManagementRole) error {
	if len(userIDs) == 0 {
		return nil
	}

	gormDB, err := gmr.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var roleTypeName string
	switch role {
	case values.GameManagementRoleAdministrator:
		roleTypeName = gameManagementRoleTypeAdministrator
	case values.GameManagementRoleCollaborator:
		roleTypeName = gameManagementRoleTypeCollaborator
	default:
		return errors.New("invalid role")
	}

	var roleType migrate.GameManagementRoleTypeTable
	err = gormDB.
		Where("name = ?", roleTypeName).
		Select("id").
		First(&roleType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	roleTypeID := roleType.ID

	gameManagementRoles := make([]*migrate.GameManagementRoleTable, 0, len(userIDs))
	for _, userID := range userIDs {
		gameManagementRoles = append(gameManagementRoles, &migrate.GameManagementRoleTable{
			GameID:     uuid.UUID(gameID),
			UserID:     uuid.UUID(userID),
			RoleTypeID: roleTypeID,
		})
	}

	err = gormDB.Create(&gameManagementRoles).Error
	if err != nil {
		return fmt.Errorf("failed to create game management roles: %w", err)
	}

	return nil
}

func (gmr *GameManagementRole) UpdateGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID, role values.GameManagementRole) error {
	gormDB, err := gmr.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var roleTypeName string
	switch role {
	case values.GameManagementRoleAdministrator:
		roleTypeName = gameManagementRoleTypeAdministrator
	case values.GameManagementRoleCollaborator:
		roleTypeName = gameManagementRoleTypeCollaborator
	default:
		return errors.New("invalid role")
	}

	var roleType migrate.GameManagementRoleTypeTable
	err = gormDB.
		Session(&gorm.Session{}).
		Where("name = ?", roleTypeName).
		Select("id").
		First(&roleType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	roleTypeID := roleType.ID

	gameManagementRole := migrate.GameManagementRoleTable{
		GameID:     uuid.UUID(gameID),
		UserID:     uuid.UUID(userID),
		RoleTypeID: roleTypeID,
	}

	result := gormDB.
		Session(&gorm.Session{}).
		Model(&gameManagementRole).
		Where("game_id = ? AND user_id = ?", uuid.UUID(gameID), uuid.UUID(userID)).
		Update("role_type_id", roleTypeID)
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to update game management role: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}

func (gmr *GameManagementRole) RemoveGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID) error {
	gormDB, err := gmr.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := gormDB.
		Where("game_id = ? AND user_id = ?", uuid.UUID(gameID), uuid.UUID(userID)).
		Delete(&migrate.GameManagementRoleTable{})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to delete game management role: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}

func (gmr *GameManagementRole) GetGameManagersByGameID(ctx context.Context, gameID values.GameID) ([]*repository.UserIDAndManagementRole, error) {
	gormDB, err := gmr.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var gameManagementRoles []migrate.GameManagementRoleTable
	err = gormDB.
		Joins("RoleTypeTable").
		Where("game_id = ?", uuid.UUID(gameID)).
		Find(&gameManagementRoles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game management roles: %w", err)
	}

	userIDAndManagementRoles := make([]*repository.UserIDAndManagementRole, 0, len(gameManagementRoles))
	for _, role := range gameManagementRoles {
		var roleType values.GameManagementRole
		switch role.RoleTypeTable.Name {
		case gameManagementRoleTypeAdministrator:
			roleType = values.GameManagementRoleAdministrator
		case gameManagementRoleTypeCollaborator:
			roleType = values.GameManagementRoleCollaborator
		default:
			return nil, errors.New("invalid role")
		}

		userIDAndManagementRoles = append(userIDAndManagementRoles, &repository.UserIDAndManagementRole{
			UserID: values.NewTrapMemberID(role.UserID),
			Role:   roleType,
		})
	}

	return userIDAndManagementRoles, nil
}

func (gmr *GameManagementRole) GetGameManagementRole(ctx context.Context, gameID values.GameID, userID values.TraPMemberID, lockType repository.LockType) (values.GameManagementRole, error) {
	gormDB, err := gmr.db.getDB(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get db: %w", err)
	}

	gormDB, err = gmr.db.setLock(gormDB, lockType)
	if err != nil {
		return 0, fmt.Errorf("failed to set lock: %w", err)
	}

	var gameManagementRole migrate.GameManagementRoleTable
	err = gormDB.
		Joins("RoleTypeTable").
		Where("game_id = ? AND user_id = ?", uuid.UUID(gameID), uuid.UUID(userID)).
		Select("RoleTypeTable.Name").
		Take(&gameManagementRole).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, repository.ErrRecordNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get game management role: %w", err)
	}

	var roleType values.GameManagementRole
	switch gameManagementRole.RoleTypeTable.Name {
	case gameManagementRoleTypeAdministrator:
		roleType = values.GameManagementRoleAdministrator
	case gameManagementRoleTypeCollaborator:
		roleType = values.GameManagementRoleCollaborator
	default:
		return 0, errors.New("invalid role")
	}

	return roleType, nil
}
