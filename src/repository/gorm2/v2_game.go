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

	gameTable := migrate.GameTable2{
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

	gameTable := migrate.GameTable2{
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
		Delete(&migrate.GameTable2{})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to remove game: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}

func (g *GameV2) GetGame(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.Game, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock type: %w", err)
	}

	var game migrate.GameTable2
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

func (g *GameV2) GetGames(ctx context.Context, limit int, offset int) ([]*domain.Game, int, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get db: %w", err)
	}

	if limit < 0 {
		return nil, 0, repository.ErrNegativeLimit
	}
	if limit == 0 && offset != 0 {
		return nil, 0, errors.New("bad limit and offset")
	}

	var games []migrate.GameTable2

	query := db.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if limit > 0 && offset > 0 {
		query = query.Offset(offset)
	}

	err = query.
		Find(&games).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games: %w", err)
	}

	gamesDomain := make([]*domain.Game, 0, len(games))
	for _, game := range games {
		gamesDomain = append(gamesDomain, domain.NewGame(
			values.NewGameIDFromUUID(game.ID),
			values.NewGameName(game.Name),
			values.NewGameDescription(game.Description),
			game.CreatedAt,
		))
	}

	var gamesNumber int64
	err = db.Table("games").
		Where("deleted_at IS NULL").
		Count(&gamesNumber).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games number: %w", err)
	}

	return gamesDomain, int(gamesNumber), nil
}

func (g *GameV2) GetGamesByUser(ctx context.Context, userID values.TraPMemberID, limit int, offset int) ([]*domain.Game, int, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get db: %w", err)
	}

	if limit < 0 {
		return nil, 0, repository.ErrNegativeLimit
	}
	if limit == 0 && offset != 0 {
		return nil, 0, errors.New("bad limit and offset")
	}

	var games []migrate.GameTable2

	tx := db.Joins("JOIN game_management_roles ON game_management_roles.game_id = games.id").
		Where("game_management_roles.user_id = ?", uuid.UUID(userID)).
		Session(&gorm.Session{})

	query := tx.
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if limit > 0 && offset > 0 {
		query = query.Offset(offset)
	}

	err = query.
		Find(&games).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games: %w", err)
	}

	gamesDomain := make([]*domain.Game, 0, len(games))
	for _, game := range games {
		gamesDomain = append(gamesDomain, domain.NewGame(
			values.NewGameIDFromUUID(game.ID),
			values.NewGameName(game.Name),
			values.NewGameDescription(game.Description),
			game.CreatedAt,
		))
	}

	var gameNumber int64

	err = tx.
		Table("games").
		Where("deleted_at IS NULL").
		Count(&gameNumber).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games number: %w", err)
	}

	return gamesDomain, int(gameNumber), nil
}
