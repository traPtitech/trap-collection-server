package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

type GameGenre struct {
	db *DB
}

func NewGameGenre(db *DB) *GameGenre {
	return &GameGenre{
		db: db,
	}
}

var _ repository.GameGenre = &GameGenre{}

func (gameGenre *GameGenre) GetGenresByGameID(ctx context.Context, gameID values.GameID) ([]*domain.GameGenre, error) {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var genres []*migrate.GameGenreTable
	err = db.
		Model(&migrate.GameGenreTable{}).
		Joins("JOIN game_genre_relations ON game_genres.id = game_genre_relations.genre_id").
		Where("game_genre_relations.game_id = ?", uuid.UUID(gameID)).
		Order("`created_at` DESC").
		Find(&genres).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*domain.GameGenre{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get GameGenre: %w", err)
	}

	result := make([]*domain.GameGenre, 0, len(genres))

	for _, genre := range genres {
		result = append(result, domain.NewGameGenre(values.GameGenreID(genre.ID), values.GameGenreName(genre.Name), genre.CreatedAt))
	}
	return result, nil
}

func (gameGenre *GameGenre) RemoveGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := db.
		Select("Games").
		Delete(&migrate.GameGenreTable{ID: uuid.UUID(gameGenreID)})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to remove game genre: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}
