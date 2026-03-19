package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameCreator struct {
	gameCreatorRepo repository.GameCreator
	gameRepository  repository.GameV2
}

func NewGameCreator(gameCreatorRepo repository.GameCreator, gameRepository repository.GameV2) *GameCreator {
	return &GameCreator{
		gameCreatorRepo: gameCreatorRepo,
		gameRepository:  gameRepository,
	}
}

func (gc *GameCreator) GetGameCreators(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorWithJobs, error) {
	_, err := gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, fmt.Errorf("get game: %w", err)
	}

	creators, err := gc.gameCreatorRepo.GetGameCreatorsByGameID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("get game creators by game id: %w", err)
	}

	return creators, nil
}

func (gc *GameCreator) GetGameCreatorJobs(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorJob, []*domain.GameCreatorCustomJob, error) {
	_, err := gc.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, nil, service.ErrInvalidGameID
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get game: %w", err)
	}

	presetJobs, err := gc.gameCreatorRepo.GetGameCreatorPresetJobs(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get game creator preset jobs: %w", err)
	}

	customJobs, err := gc.gameCreatorRepo.GetGameCreatorCustomJobsByGameID(ctx, gameID)
	if err != nil {
		return nil, nil, fmt.Errorf("get game creator custom jobs by game id: %w", err)
	}

	return presetJobs, customJobs, nil
}
