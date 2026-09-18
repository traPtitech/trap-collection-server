package service

import "github.com/traPtitech/trap-collection-server/src/domain/values"

type FeedbackQuestionInput struct {
	ID           *values.FeedbackQuestionID
	QuestionText values.FeedbackQuestionText
	AnswerType   values.FeedbackAnswerType
}
