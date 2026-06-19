package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameFeedback struct {
	gameRepository         repository.GameV2
	gameFeedbackRepository repository.GameFeedback
}

func NewGameFeedback(
	gameRepository repository.GameV2,
	gameFeedbackRepository repository.GameFeedback,
) *GameFeedback {
	return &GameFeedback{
		gameRepository:         gameRepository,
		gameFeedbackRepository: gameFeedbackRepository,
	}
}

func (g *GameFeedback) GetFeedbackConfig(ctx context.Context, gameID values.GameID) (bool, error) {
	_, err := g.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return false, service.ErrInvalidGame
	}
	if err != nil {
		return false, fmt.Errorf("failed to get game: %w", err)
	}

	enabled, err := g.gameFeedbackRepository.GetFeedbackConfig(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get game feedback config: %w", err)
	}

	return enabled, nil
}
