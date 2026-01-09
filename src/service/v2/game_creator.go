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

type GameCreatorService struct {
	gameCreatorRepo repository.GameCreator
	gameRepository  repository.GameV2
}

func NewGameCreatorService(gameCreatorRepo repository.GameCreator, gameRepository repository.GameV2) *GameCreatorService {
	return &GameCreatorService{
		gameCreatorRepo: gameCreatorRepo,
		gameRepository:  gameRepository,
	}
}

func (gc *GameCreatorService) GetGameCreators(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorWithJobs, error) {
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
