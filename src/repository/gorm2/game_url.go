package gorm2

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameURL struct {
	db *DB
}

func NewGameURL(db *DB) *GameURL {
	return &GameURL{
		db: db,
	}
}

func (gu *GameURL) SaveGameURL(ctx context.Context, gameVersionID values.GameVersionID, gameURL *domain.GameURL) error {
	db, err := gu.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	err = db.Create(&GameURLTable{
		ID:            uuid.UUID(gameURL.GetID()),
		GameVersionID: uuid.UUID(gameVersionID),
		URL:           (*url.URL)(gameURL.GetLink()).String(),
	}).Error
	if err != nil {
		return fmt.Errorf("failed to create game url: %w", err)
	}

	return nil
}
