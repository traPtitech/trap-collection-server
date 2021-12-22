package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
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
