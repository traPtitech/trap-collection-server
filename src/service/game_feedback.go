package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type GameFeedback interface {
	GetFeedbackConfig(ctx context.Context, gameID values.GameID) (bool, error)
	GetFeedbackQuestions(ctx context.Context, gameID values.GameID) ([]*domain.FeedbackQuestion, error)
	PutFeedbackQuestions(ctx context.Context, gameID values.GameID, inputs []FeedbackQuestionInput) ([]*domain.FeedbackQuestion, error)
	GetGameFeedbacks(ctx context.Context, gameID values.GameID, limit, offset int) ([]*GameFeedbackDetail, int, error)
	GetGameVersionFeedbacks(ctx context.Context, gameID values.GameID, gameVersionID values.GameVersionID, limit, offset int) ([]*GameFeedbackDetail, int, error)
}

type GameFeedbackDetail struct {
	Feedback *domain.GameFeedback
	Answers  []*GameFeedbackAnswerDetail
}

type GameFeedbackAnswerDetail struct {
	Answer       *domain.GameFeedbackAnswer
	QuestionText values.FeedbackQuestionText
	AnswerType   values.FeedbackAnswerType
}
