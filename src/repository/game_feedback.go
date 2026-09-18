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
	CreateFeedbackQuestions(ctx context.Context, questions []*domain.FeedbackQuestion) error
	UpdateFeedbackQuestions(ctx context.Context, questions []*domain.FeedbackQuestion) error
	ArchiveFeedbackQuestions(ctx context.Context, ids []values.FeedbackQuestionID) error
}
