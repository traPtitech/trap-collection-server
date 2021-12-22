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

func (g *Game) DeleteGame(ctx context.Context, id values.GameID) error {
	err := g.gameRepository.RemoveGame(ctx, id)
	if errors.Is(err, repository.ErrNoRecordDeleted) {
		return service.ErrNoGame
	}
	if err != nil {
		return fmt.Errorf("failed to delete game: %w", err)
	}

	return nil
}

func (g *Game) GetGame(ctx context.Context, id values.GameID) (*service.GameInfo, error) {
	game, err := g.gameRepository.GetGame(ctx, id, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGame
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	gameVersion, err := g.gameVersionRepository.GetLatestGameVersion(ctx, id, repository.LockTypeNone)
	if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to get game version: %w", err)
	}

	if errors.Is(err, repository.ErrRecordNotFound) {
		gameVersion = nil
	}

	return &service.GameInfo{
		Game:          game,
		LatestVersion: gameVersion,
	}, nil
}

func (g *Game) GetGames(ctx context.Context) ([]*service.GameInfo, error) {
	games, err := g.gameRepository.GetGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get games: %w", err)
	}

	if len(games) == 0 {
		return []*service.GameInfo{}, nil
	}

	gameIDs := make([]values.GameID, 0, len(games))
	for _, game := range games {
		gameIDs = append(gameIDs, game.GetID())
	}

	gameVersions, err := g.gameVersionRepository.GetLatestGameVersionsByGameIDs(ctx, gameIDs, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game versions: %w", err)
	}

	var gameInfos []*service.GameInfo
	for _, game := range games {
		gameVersion, ok := gameVersions[game.GetID()]
		if !ok {
			gameVersion = nil
		}

		gameInfos = append(gameInfos, &service.GameInfo{
			Game:          game,
			LatestVersion: gameVersion,
		})
	}

	return gameInfos, nil
}

func (g *Game) GetMyGames(ctx context.Context, session *domain.OIDCSession) ([]*service.GameInfo, error) {
	user, err := g.userUtils.getMe(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	games, err := g.gameRepository.GetGamesByUser(ctx, user.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to get game ids: %w", err)
	}

	gameIDs := make([]values.GameID, 0, len(games))
	for _, game := range games {
		gameIDs = append(gameIDs, game.GetID())
	}

	if len(games) == 0 {
		return []*service.GameInfo{}, nil
	}

	gameVersions, err := g.gameVersionRepository.GetLatestGameVersionsByGameIDs(ctx, gameIDs, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get game versions: %w", err)
	}

	var gameInfos []*service.GameInfo
	for _, game := range games {
		gameVersion, ok := gameVersions[game.GetID()]
		if !ok {
			gameVersion = nil
		}

		gameInfos = append(gameInfos, &service.GameInfo{
			Game:          game,
			LatestVersion: gameVersion,
		})
	}

	return gameInfos, nil
}
