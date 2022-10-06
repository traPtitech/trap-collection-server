package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

type GameV2 struct {
	db *DB
}

func NewGameV2(db *DB) *GameV2 {
	return &GameV2{
		db: db,
	}
}

func (g *GameV2) SaveGame(ctx context.Context, game *domain.Game) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	gameTable := migrate.GameTable{
		ID:          uuid.UUID(game.GetID()),
		Name:        string(game.GetName()),
		Description: string(game.GetDescription()),
		CreatedAt:   game.GetCreatedAt(),
	}

	err = db.Create(&gameTable).Error
	if err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	return nil
}

func (g *GameV2) UpdateGame(ctx context.Context, game *domain.Game) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	gameTable := migrate.GameTable{
		Name:        string(game.GetName()),
		Description: string(game.GetDescription()),
	}

	result := db.
		Where("id = ?", uuid.UUID(game.GetID())).
		Updates(gameTable)
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}

func (g *GameV2) RemoveGame(ctx context.Context, gameID values.GameID) error {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := db.
		Where("id = ?", uuid.UUID(gameID)).
		Delete(&migrate.GameTable{})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to remove game: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}
