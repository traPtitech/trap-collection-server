package gorm2

import (
	"context"
	"fmt"

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
