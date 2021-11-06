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

type GameVersion struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
}

func NewGameVersion(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
) *GameVersion {
	return &GameVersion{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
	}
}

func (gv *GameVersion) CreateGameVersion(ctx context.Context, gameID values.GameID, name values.GameVersionName, description values.GameVersionDescription) (*domain.GameVersion, error) {
	var gameVersion *domain.GameVersion
	err := gv.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gv.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		gameVersion = domain.NewGameVersion(
			values.NewGameVersionID(),
			name,
			description,
			time.Now(),
		)

		err = gv.gameVersionRepository.CreateGameVersion(ctx, gameID, gameVersion)
		if err != nil {
			return fmt.Errorf("failed to create game version: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return gameVersion, nil
}

func (gv *GameVersion) GetGameVersions(ctx context.Context, gameID values.GameID) ([]*domain.GameVersion, error) {
	var gameVersions []*domain.GameVersion
	err := gv.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gv.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		gameVersions, err = gv.gameVersionRepository.GetGameVersions(ctx, gameID)
		if err != nil {
			return fmt.Errorf("failed to get game versions: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return gameVersions, nil
}
