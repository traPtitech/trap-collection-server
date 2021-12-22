package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Game struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
	userUtils             *UserUtils
}

func NewGame(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
	userUtils *UserUtils,
) *Game {
	return &Game{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
		userUtils:             userUtils,
	}
}

func (g *Game) CreateGame(ctx context.Context, name values.GameName, description values.GameDescription) (*domain.Game, error) {
	game := domain.NewGame(
		values.NewGameID(),
		name,
		description,
		time.Now(),
	)

	err := g.gameRepository.SaveGame(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	return game, nil
}

func (g *Game) UpdateGame(ctx context.Context, gameID values.GameID, name values.GameName, description values.GameDescription) (*domain.Game, error) {
	var game *domain.Game
	err := g.db.Transaction(ctx, nil, func(ctx context.Context) error {
		var err error
		game, err = g.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGame
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		// 変更がなければ何もしない
		if game.GetName() == name && game.GetDescription() == description {
			return nil
		}

		game.SetName(name)
		game.SetDescription(description)

		err = g.gameRepository.UpdateGame(ctx, game)
		if err != nil {
			return fmt.Errorf("failed to save game: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return game, nil
}
