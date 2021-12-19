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

type GameURL struct {
	db                    repository.DB
	gameRepository        repository.Game
	gameVersionRepository repository.GameVersion
	gameURLRepository     repository.GameURL
}

func NewGameURL(
	db repository.DB,
	gameRepository repository.Game,
	gameVersionRepository repository.GameVersion,
	gameURLRepository repository.GameURL,
) *GameURL {
	return &GameURL{
		db:                    db,
		gameRepository:        gameRepository,
		gameVersionRepository: gameVersionRepository,
		gameURLRepository:     gameURLRepository,
	}
}

func (gu *GameURL) SaveGameURL(ctx context.Context, gameID values.GameID, link values.GameURLLink) (*domain.GameURL, error) {
	var gameURL *domain.GameURL
	err := gu.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := gu.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGameID
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		gameVersion, err := gu.gameVersionRepository.GetLatestGameVersion(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrNoGameVersion
		}
		if err != nil {
			return fmt.Errorf("failed to get latest game version: %w", err)
		}

		_, err = gu.gameURLRepository.GetGameURL(ctx, gameVersion.GetID())
		if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
			return fmt.Errorf("failed to get game url: %w", err)
		}

		if err == nil {
			return service.ErrGameURLAlreadyExists
		}

		gameURLID := values.NewGameURLID()
		gameURL = domain.NewGameURL(gameURLID, link, time.Now())

		err = gu.gameURLRepository.SaveGameURL(ctx, gameVersion.GetID(), gameURL)
		if err != nil {
			return fmt.Errorf("failed to save game url: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return gameURL, nil
}

func (gu *GameURL) GetGameURL(ctx context.Context, gameID values.GameID) (*domain.GameURL, error) {
	_, err := gu.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	gameVersion, err := gu.gameVersionRepository.GetLatestGameVersion(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGameVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest game version: %w", err)
	}

	gameURL, err := gu.gameURLRepository.GetGameURL(ctx, gameVersion.GetID())
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrNoGameURL
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game url: %w", err)
	}

	return gameURL, nil
}
