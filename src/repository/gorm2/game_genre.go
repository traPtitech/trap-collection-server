package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
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

func (gameGenre *GameGenre) GetGameGenresWithNames(ctx context.Context, gameGenreNames []values.GameGenreName) ([]*domain.GameGenre, error) {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var genres []migrate.GameGenreTable
	result := db.
		Where("name IN ?", gameGenreNames).
		Find(&genres)
	if result.RowsAffected == 0 {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game genres with names: %w", err)
	}

	resultGenres := make([]*domain.GameGenre, 0, len(genres))
	for _, genre := range genres {
		resultGenres = append(resultGenres, domain.NewGameGenre(values.GameGenreID(genre.ID), values.GameGenreName(genre.Name), genre.CreatedAt))
	}

	return resultGenres, nil
}

// // SaveGameGenres
// // ゲームジャンルを作成する。
// // 名前が重複するゲームジャンルが1つでも存在するとき、ErrDuplicatedUniqueKeyを返す。
func (gameGenre *GameGenre) SaveGameGenres(ctx context.Context, gameGenres []*domain.GameGenre) error {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	genres := make([]migrate.GameGenreTable, 0, len(gameGenres))
	for i := range gameGenres {
		genres = append(genres, migrate.GameGenreTable{
			ID:        uuid.UUID(gameGenres[i].GetID()),
			Name:      string(gameGenres[i].GetName()),
			CreatedAt: gameGenres[i].GetCreatedAt(),
		})
	}

	err = db.Create(&genres).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == 1062 {
			return repository.ErrDuplicatedUniqueKey
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// RegisterGenresToGame
// ゲームにゲームジャンルを登録する。
func (gameGenre *GameGenre) RegisterGenresToGame(ctx context.Context, gameID values.GameID, gameGenreIDs []values.GameGenreID) error {
	db, err := gameGenre.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var game migrate.GameTable2

	if err = db.First(&game, uuid.UUID(gameID)).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return repository.ErrRecordNotFound
	} else if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	genres := make([]uuid.UUID, 0, len(gameGenreIDs))
	for _, genre := range gameGenreIDs {
		genres = append(genres, uuid.UUID(genre))
	}

	var gameGenres []migrate.GameGenreTable
	err = db.Find(&gameGenres, genres).Error
	if err != nil {
		return fmt.Errorf("failed to get game genres: %w", err)
	}

	if len(gameGenres) != len(gameGenreIDs) {
		return repository.ErrIncludeInvalidArgs
	}

	err = db.Model(&game).Association("GameGenres").Replace(gameGenres)
	if err != nil {
		return fmt.Errorf("failed to register genres to game: %w", err)
	}

	return nil
}
