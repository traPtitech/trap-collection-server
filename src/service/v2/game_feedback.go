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

type GameFeedback struct {
	gameRepository         repository.GameV2
	gameFeedbackRepository repository.GameFeedback
	gameVersionRepository  repository.GameVersionV2
}

func NewGameFeedback(
	gameRepository repository.GameV2,
	gameFeedbackRepository repository.GameFeedback,
	gameVersionRepository repository.GameVersionV2,
) *GameFeedback {
	return &GameFeedback{
		gameRepository:         gameRepository,
		gameFeedbackRepository: gameFeedbackRepository,
		gameVersionRepository:  gameVersionRepository,
	}
}

func (g *GameFeedback) GetFeedbackQuestions(ctx context.Context, gameID values.GameID) ([]*domain.FeedbackQuestion, error) {
	_, err := g.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidGame
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	questions, err := g.gameFeedbackRepository.GetFeedbackQuestions(ctx, gameID, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback questions: %w", err)
	}
	return questions, nil
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
