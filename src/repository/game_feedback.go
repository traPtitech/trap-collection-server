package repository

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFeedback interface {
	GetFeedbackConfig(ctx context.Context, gameID values.GameID, lockType LockType) (bool, error)
	GetFeedbackQuestions(ctx context.Context, gameID values.GameID, lockType LockType) ([]*domain.FeedbackQuestion, error)
	GetFeedbackQuestionsIncludingArchived(ctx context.Context, gameID values.GameID, lockType LockType) ([]*domain.FeedbackQuestion, error)
	GetGameFeedbacksByGameID(ctx context.Context, gameID values.GameID, limit, offset int) ([]*GameFeedbackWithAnswers, int, error)
}

type GameFeedbackWithAnswers struct {
	Feedback *domain.GameFeedback
	Answers  []*domain.GameFeedbackAnswer
}
