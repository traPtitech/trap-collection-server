package gorm2

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
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

func (gu *GameURL) GetGameURL(ctx context.Context, gameVersionID values.GameVersionID) (*domain.GameURL, error) {
	db, err := gu.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var gameURLTable GameURLTable
	err = db.
		Where("game_version_id = ?", uuid.UUID(gameVersionID)).
		Take(&gameURLTable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game url: %w", err)
	}

	urlGameLink, err := url.Parse(gameURLTable.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse game url: %w", err)
	}

	return domain.NewGameURL(
		values.NewGameURLIDFromUUID(gameURLTable.ID),
		values.NewGameURLLink(urlGameLink),
	), nil
}
