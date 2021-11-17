package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVersion struct {
	db *DB
}

func NewGameVersion(db *DB) *GameVersion {
	return &GameVersion{
		db: db,
	}
}

func (gv *GameVersion) CreateGameVersion(ctx context.Context, gameID values.GameID, version *domain.GameVersion) error {
	gormDB, err := gv.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gorm DB: %w", err)
	}

	gameVersion := GameVersionTable{
		ID:          uuid.UUID(version.GetID()),
		GameID:      uuid.UUID(gameID),
		Name:        string(version.GetName()),
		Description: string(version.GetDescription()),
		CreatedAt:   version.GetCreatedAt(),
	}

	err = gormDB.Create(&gameVersion).Error
	if err != nil {
		return fmt.Errorf("failed to create game version: %w", err)
	}

	return nil
}
