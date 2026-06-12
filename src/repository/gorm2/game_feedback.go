package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

type GameFeedback struct {
	db *DB
}

var _ repository.GameFeedback = (*GameFeedback)(nil)

func NewGameFeedback(db *DB) *GameFeedback {
	return &GameFeedback{
		db: db,
	}
}

func (g *GameFeedback) GetFeedbackConfig(ctx context.Context, gameID values.GameID, lockType repository.LockType) (bool, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return false, fmt.Errorf("failed to set lock: %w", err)
	}

	var config schema.GameFeedbackConfigTable
	err = db.Where("game_id = ?", uuid.UUID(gameID)).Take(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, repository.ErrRecordNotFound
	}
	if err != nil {
		return false, fmt.Errorf("failed to get game feedback config: %w", err)
	}

	return config.Enabled, nil
}
