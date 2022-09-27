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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
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

	err = db.Create(&migrate.GameURLTable{
		ID:            uuid.UUID(gameURL.GetID()),
		GameVersionID: uuid.UUID(gameVersionID),
		URL:           (*url.URL)(gameURL.GetLink()).String(),
		CreatedAt:     gameURL.GetCreatedAt(),
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

	var gameURLTable migrate.GameURLTable
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
		gameURLTable.CreatedAt,
	), nil
}

func (gu *GameURL) GetGameURLsByGameVersionIDs(ctx context.Context, gameVersionIDs []values.GameID, lockType repository.LockType) (map[values.GameVersionID]*domain.GameURL, error) {
	db, err := gu.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gorm DB: %w", err)
	}

	db, err = gu.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	uuidGameVersionIDs := make([]uuid.UUID, 0, len(gameVersionIDs))
	for _, gameID := range gameVersionIDs {
		uuidGameVersionIDs = append(uuidGameVersionIDs, uuid.UUID(gameID))
	}

	var gameURLTables []migrate.gameURLTable
	err = db.
		Where("game_version_id in (?)", uuidGameVersionIDs).
		Find(&gameURLTables).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get latest game versions: %w", err)
	}

	gameURLs := make(map[values.GameVersionID]*domain.GameURL, len(gameURLTables))
	for _, gameURL := range gameURLTables {
		gameVersionID := values.NewGameVersionIDFromUUID(gameURL.GameVersionID)
		gameURLID := values.NewGameURLIDFromUUID(gameURL.ID)
		urlGameURLLink, err := url.Parse(gameURL.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse game url: %w", err)
		}
		gameURLLink := values.NewGameURLLink(urlGameURLLink)

		gameURLs[gameVersionID] = domain.NewGameURL(
			gameURLID,
			gameURLLink,
			gameURL.CreatedAt,
		)
	}

	return gameURLs, nil
}
