package v2

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

type GameFeedback struct {
	db                     repository.DB
	gameRepository         repository.GameV2
	gameFeedbackRepository repository.GameFeedback
}

func NewGameFeedback(
	db repository.DB,
	gameRepository repository.GameV2,
	gameFeedbackRepository repository.GameFeedback,
) *GameFeedback {
	return &GameFeedback{
		db:                     db,
		gameRepository:         gameRepository,
		gameFeedbackRepository: gameFeedbackRepository,
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

func (g *GameFeedback) PutFeedbackQuestions(ctx context.Context, gameID values.GameID, inputs []service.FeedbackQuestionInput) ([]*domain.FeedbackQuestion, error) {
	questions := make([]*domain.FeedbackQuestion, 0, len(inputs))
	err := g.db.Transaction(ctx, nil, func(ctx context.Context) error {
		_, err := g.gameRepository.GetGame(ctx, gameID, repository.LockTypeRecord)
		if errors.Is(err, repository.ErrRecordNotFound) {
			return service.ErrInvalidGame
		}
		if err != nil {
			return fmt.Errorf("failed to get game: %w", err)
		}

		existing, err := g.gameFeedbackRepository.GetFeedbackQuestions(ctx, gameID, repository.LockTypeNone)
		if err != nil {
			return fmt.Errorf("failed to get feedback questions: %w", err)
		}
		existingByID := make(map[values.FeedbackQuestionID]*domain.FeedbackQuestion, len(existing))
		for _, question := range existing {
			existingByID[question.GetID()] = question
		}

		seen := make(map[values.FeedbackQuestionID]struct{}, len(inputs))
		newQuestions := make([]*domain.FeedbackQuestion, 0, len(inputs))
		updatedQuestions := make([]*domain.FeedbackQuestion, 0, len(inputs))
		for order, input := range inputs {
			if err := input.QuestionText.Validate(); err != nil {
				return fmt.Errorf("%w: %w", service.ErrInvalidFeedbackQuestion, err)
			}
			if input.AnswerType != values.FeedbackAnswerTypeYesNo && input.AnswerType != values.FeedbackAnswerTypeFiveScale {
				return service.ErrInvalidFeedbackAnswerType
			}

			questionOrder := values.NewFeedbackQuestionOrder(order)
			if input.ID == nil {
				question := domain.NewFeedbackQuestion(
					values.NewFeedbackQuestionID(), gameID, input.QuestionText, input.AnswerType,
					questionOrder, time.Now(), nil,
				)
				questions = append(questions, question)
				newQuestions = append(newQuestions, question)
				continue
			}

			if _, ok := seen[*input.ID]; ok {
				return service.ErrDuplicateFeedbackQuestion
			}
			seen[*input.ID] = struct{}{}

			existingQuestion, ok := existingByID[*input.ID]
			if !ok {
				return service.ErrInvalidFeedbackQuestion
			}
			question := domain.NewFeedbackQuestion(
				existingQuestion.GetID(), gameID, input.QuestionText, input.AnswerType,
				questionOrder, existingQuestion.GetCreatedAt(), nil,
			)
			questions = append(questions, question)
			updatedQuestions = append(updatedQuestions, question)
		}

		archiveIDs := make([]values.FeedbackQuestionID, 0, len(existing))
		for _, question := range existing {
			if _, ok := seen[question.GetID()]; !ok {
				archiveIDs = append(archiveIDs, question.GetID())
			}
		}

		if err := g.gameFeedbackRepository.UpdateFeedbackQuestions(ctx, updatedQuestions); err != nil {
			return fmt.Errorf("failed to update feedback questions: %w", err)
		}
		if err := g.gameFeedbackRepository.CreateFeedbackQuestions(ctx, newQuestions); err != nil {
			return fmt.Errorf("failed to create feedback questions: %w", err)
		}
		if err := g.gameFeedbackRepository.ArchiveFeedbackQuestions(ctx, archiveIDs); err != nil {
			return fmt.Errorf("failed to archive feedback questions: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
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
