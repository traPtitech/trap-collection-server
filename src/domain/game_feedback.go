package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameFeedbackPreference struct {
	gameID  values.GameID
	enabled bool
}

func NewGameFeedbackPreference(
	gameID values.GameID,
	enabled bool,
) *GameFeedbackPreference {
	return &GameFeedbackPreference{
		gameID:  gameID,
		enabled: enabled,
	}
}

func (c *GameFeedbackPreference) GetGameID() values.GameID {
	return c.gameID
}

func (c *GameFeedbackPreference) IsEnabled() bool {
	return c.enabled
}

type FeedbackQuestion struct {
	id            values.FeedbackQuestionID
	gameID        values.GameID
	questionText  values.FeedbackQuestionText
	answerType    values.FeedbackAnswerType
	questionOrder values.FeedbackQuestionOrder
	createdAt     time.Time
	archivedAt    *time.Time
}

func NewFeedbackQuestion(
	id values.FeedbackQuestionID,
	gameID values.GameID,
	questionText values.FeedbackQuestionText,
	answerType values.FeedbackAnswerType,
	questionOrder values.FeedbackQuestionOrder,
	createdAt time.Time,
	archivedAt *time.Time,
) *FeedbackQuestion {
	return &FeedbackQuestion{
		id:            id,
		gameID:        gameID,
		questionText:  questionText,
		answerType:    answerType,
		questionOrder: questionOrder,
		createdAt:     createdAt,
		archivedAt:    archivedAt,
	}
}

func (q *FeedbackQuestion) GetID() values.FeedbackQuestionID {
	return q.id
}

func (q *FeedbackQuestion) GetGameID() values.GameID {
	return q.gameID
}

func (q *FeedbackQuestion) GetQuestionText() values.FeedbackQuestionText {
	return q.questionText
}

func (q *FeedbackQuestion) GetAnswerType() values.FeedbackAnswerType {
	return q.answerType
}

func (q *FeedbackQuestion) GetQuestionOrder() values.FeedbackQuestionOrder {
	return q.questionOrder
}

func (q *FeedbackQuestion) GetCreatedAt() time.Time {
	return q.createdAt
}

func (q *FeedbackQuestion) GetArchivedAt() *time.Time {
	return q.archivedAt
}

func (q *FeedbackQuestion) IsArchived() bool {
	return q.archivedAt != nil
}

type GameFeedback struct {
	id            values.GameFeedbackID
	gameVersionID values.GameVersionID
	comment       *values.FeedbackComment
	createdAt     time.Time
}

func NewGameFeedback(
	id values.GameFeedbackID,
	gameVersionID values.GameVersionID,
	comment *values.FeedbackComment,
	createdAt time.Time,
) *GameFeedback {
	return &GameFeedback{
		id:            id,
		gameVersionID: gameVersionID,
		comment:       comment,
		createdAt:     createdAt,
	}
}

func (f *GameFeedback) GetID() values.GameFeedbackID {
	return f.id
}

func (f *GameFeedback) GetGameVersionID() values.GameVersionID {
	return f.gameVersionID
}

func (f *GameFeedback) GetComment() *values.FeedbackComment {
	return f.comment
}

func (f *GameFeedback) GetCreatedAt() time.Time {
	return f.createdAt
}

type GameFeedbackAnswer struct {
	id         values.GameFeedbackAnswerID
	feedbackID values.GameFeedbackID
	questionID values.FeedbackQuestionID
	// answer は質問の回答種別ごとに以下の値を取る。
	// FeedbackAnswerTypeYesNo: 0 = No, 1 = Yes
	// FeedbackAnswerTypeFiveScale: 1〜5
	answer int
}

func NewGameFeedbackAnswer(
	id values.GameFeedbackAnswerID,
	feedbackID values.GameFeedbackID,
	questionID values.FeedbackQuestionID,
	answer int,
) *GameFeedbackAnswer {
	return &GameFeedbackAnswer{
		id:         id,
		feedbackID: feedbackID,
		questionID: questionID,
		answer:     answer,
	}
}

func (a *GameFeedbackAnswer) GetID() values.GameFeedbackAnswerID {
	return a.id
}

func (a *GameFeedbackAnswer) GetFeedbackID() values.GameFeedbackID {
	return a.feedbackID
}

func (a *GameFeedbackAnswer) GetQuestionID() values.FeedbackQuestionID {
	return a.questionID
}

func (a *GameFeedbackAnswer) GetAnswer() int {
	return a.answer
}
