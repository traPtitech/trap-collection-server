package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
)

const (
	gameManagementRoleTypeAdministrator = "administrator"
	gameManagementRoleTypeCollaborator  = "collaborator"
)

type GameManagementRole struct {
	db *DB
}

func NewGameManagementRole(db *DB) (*GameManagementRole, error) {
	ctx := context.Background()

	gormDB, err := db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	err = setupRoleTypeTable(gormDB)
	if err != nil {
		return nil, fmt.Errorf("failed to setup role type table: %w", err)
	}

	return &GameManagementRole{
		db: db,
	}, nil
}

func setupRoleTypeTable(db *gorm.DB) error {
	roleTypes := []GameManagementRoleTypeTable{
		{
			Name: gameManagementRoleTypeAdministrator,
		},
		{
			Name: gameManagementRoleTypeCollaborator,
		},
	}

	for _, roleType := range roleTypes {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", roleType.Name).
			FirstOrCreate(&roleType).Error
		if err != nil {
			return fmt.Errorf("failed to create role type: %w", err)
		}
	}

	return nil
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

	var roleType GameManagementRoleTypeTable
	err = gormDB.
		Where("name = ?", roleTypeName).
		Select("id").
		First(&roleType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	roleTypeID := roleType.ID

	gameManagementRoles := make([]*GameManagementRoleTable, 0, len(userIDs))
	for _, userID := range userIDs {
		gameManagementRoles = append(gameManagementRoles, &GameManagementRoleTable{
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

	var roleType GameManagementRoleTypeTable
	err = gormDB.
		Session(&gorm.Session{}).
		Where("name = ?", roleTypeName).
		Select("id").
		First(&roleType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	roleTypeID := roleType.ID

	gameManagementRole := GameManagementRoleTable{
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
		Delete(&GameManagementRoleTable{})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to delete game management role: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}
