package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
)

type Game struct {
	db *DB
}

func NewGame(db *DB) *Game {
	return &Game{
		db: db,
	}
}

func (g *Game) GetGame(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.Game, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock type: %w", err)
	}

	var game GameTable
	err = db.
		Where("id = ?", uuid.UUID(gameID)).
		Take(&game).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return domain.NewGame(
		values.NewGameIDFromUUID(game.ID),
		values.NewGameName(game.Name),
		values.NewGameDescription(game.Description),
		game.CreatedAt,
	), nil
}
