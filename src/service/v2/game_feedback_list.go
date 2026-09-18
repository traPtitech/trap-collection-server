package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

func (g *GameFeedback) GetGameFeedbacks(ctx context.Context, gameID values.GameID, limit, offset int) ([]*service.GameFeedbackDetail, int, error) {
	if err := validateGameFeedbackPagination(limit, offset); err != nil {
		return nil, 0, err
	}

	if err := g.validateGame(ctx, gameID); err != nil {
		return nil, 0, err
	}

	feedbacks, total, err := g.gameFeedbackRepository.GetGameFeedbacksByGameID(ctx, gameID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get game feedbacks: %w", err)
	}

	return g.feedbackDetails(ctx, gameID, feedbacks, total)
}

func (g *GameFeedback) GetGameVersionFeedbacks(ctx context.Context, gameID values.GameID, gameVersionID values.GameVersionID, limit, offset int) ([]*service.GameFeedbackDetail, int, error) {
	if err := validateGameFeedbackPagination(limit, offset); err != nil {
		return nil, 0, err
	}
	if err := g.validateGame(ctx, gameID); err != nil {
		return nil, 0, err
	}
	gameVersion, err := g.gameVersionRepository.GetGameVersionByID(ctx, gameVersionID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, 0, service.ErrInvalidGameVersion
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get game version: %w", err)
	}
	if gameVersion.GameID != gameID {
		return nil, 0, service.ErrInvalidGameVersion
	}
	feedbacks, total, err := g.gameFeedbackRepository.GetGameFeedbacksByGameVersionID(ctx, gameVersionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get game version feedbacks: %w", err)
	}
	return g.feedbackDetails(ctx, gameID, feedbacks, total)
}

func (g *GameFeedback) validateGame(ctx context.Context, gameID values.GameID) error {
	_, err := g.gameRepository.GetGame(ctx, gameID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return service.ErrInvalidGame
	}
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}
	return nil
}

func (g *GameFeedback) feedbackDetails(ctx context.Context, gameID values.GameID, feedbacks []*repository.GameFeedbackWithAnswers, total int) ([]*service.GameFeedbackDetail, int, error) {
	questions, err := g.gameFeedbackRepository.GetFeedbackQuestionsIncludingArchived(ctx, gameID, repository.LockTypeNone)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get feedback questions: %w", err)
	}

	questionsByID := make(map[values.FeedbackQuestionID]struct {
		text       values.FeedbackQuestionText
		answerType values.FeedbackAnswerType
	}, len(questions))
	for _, question := range questions {
		questionsByID[question.GetID()] = struct {
			text       values.FeedbackQuestionText
			answerType values.FeedbackAnswerType
		}{
			text:       question.GetQuestionText(),
			answerType: question.GetAnswerType(),
		}
	}

	result := make([]*service.GameFeedbackDetail, 0, len(feedbacks))
	for _, feedback := range feedbacks {
		answers := make([]*service.GameFeedbackAnswerDetail, 0, len(feedback.Answers))
		for _, answer := range feedback.Answers {
			question, ok := questionsByID[answer.GetQuestionID()]
			if !ok {
				continue
			}
			answers = append(answers, &service.GameFeedbackAnswerDetail{
				Answer:       answer,
				QuestionText: question.text,
				AnswerType:   question.answerType,
			})
		}
		result = append(result, &service.GameFeedbackDetail{
			Feedback: feedback.Feedback,
			Answers:  answers,
		})
	}

	return result, total, nil
}

func validateGameFeedbackPagination(limit, offset int) error {
	if limit < 1 || limit > 100 || offset < 0 {
		return service.ErrInvalidLimit
	}
	return nil
}
