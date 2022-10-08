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

func (g *GameV2) SaveGameV2(ctx context.Context, game *domain.Game) error {
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

func (g *GameV2) UpdateGameV2(ctx context.Context, game *domain.Game) error {
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

func (g *GameV2) RemoveGameV2(ctx context.Context, gameID values.GameID) error {
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

func (g *GameV2) GetGameV2(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.Game, error) {
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

func (g *GameV2) GetGamesV2(ctx context.Context, limit int, offset int) ([]*domain.Game, int, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get db: %w", err)
	}

	var allGames []migrate.GameTable2
	err = db.
		Order("created_at DESC").
		Find(&allGames).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games: %w", err)
	}

	games := make([]migrate.GameTable2, 0, len(allGames))

	if limit == -1 {
		games = allGames[offset:]
	} else if limit < -1 {
		return nil, 0, repository.ErrNegativeLimit
	} else {
		games = allGames[offset : offset+limit]
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

	return gamesDomain, len(allGames), nil
}

func (g *GameV2) GetGamesByUserV2(ctx context.Context, userID values.TraPMemberID, limit int, offset int) ([]*domain.Game, int, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get db: %w", err)
	}

	var allUserGames []migrate.GameTable2
	err = db.
		Joins("JOIN game_management_roles ON game_management_roles.game_id = games.id").
		Where("game_management_roles.user_id = ?", uuid.UUID(userID)).
		Order("created_at DESC").
		Find(&allUserGames).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get games: %w", err)
	}

	games := make([]migrate.GameTable2, 0, len(allUserGames))
	if limit == -1 {
		games = allUserGames[offset:]
	} else if limit < -1 {
		return nil, 0, repository.ErrNegativeLimit
	} else {
		games = allUserGames[offset : offset+limit]
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

	return gamesDomain, len(allUserGames), nil
}
