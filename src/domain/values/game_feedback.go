package values

import (
	"errors"
	"unicode/utf8"

	"github.com/google/uuid"
)

type (
	FeedbackQuestionID    uuid.UUID
	GameFeedbackID        uuid.UUID
	GameFeedbackAnswerID  uuid.UUID
	FeedbackAnswerType    int
	FeedbackQuestionText  string
	FeedbackQuestionOrder int
	FeedbackComment       string
)

func NewFeedbackQuestionID() FeedbackQuestionID {
	return FeedbackQuestionID(uuid.New())
}

func NewFeedbackQuestionIDFromUUID(id uuid.UUID) FeedbackQuestionID {
	return FeedbackQuestionID(id)
}

func (id FeedbackQuestionID) String() string {
	return uuid.UUID(id).String()
}

func (id FeedbackQuestionID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func NewGameFeedbackID() GameFeedbackID {
	return GameFeedbackID(uuid.New())
}

func NewGameFeedbackIDFromUUID(id uuid.UUID) GameFeedbackID {
	return GameFeedbackID(id)
}

func (id GameFeedbackID) String() string {
	return uuid.UUID(id).String()
}

func (id GameFeedbackID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func NewGameFeedbackAnswerID() GameFeedbackAnswerID {
	return GameFeedbackAnswerID(uuid.New())
}

func NewGameFeedbackAnswerIDFromUUID(id uuid.UUID) GameFeedbackAnswerID {
	return GameFeedbackAnswerID(id)
}

func (id GameFeedbackAnswerID) String() string {
	return uuid.UUID(id).String()
}

func (id GameFeedbackAnswerID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

const (
	// FeedbackAnswerTypeYesNo は Yes/No で答える質問。回答は 0 = No, 1 = Yes。
	FeedbackAnswerTypeYesNo FeedbackAnswerType = iota
	// FeedbackAnswerTypeFiveScale は 1〜5 の5段階で答える質問。
	FeedbackAnswerTypeFiveScale
)

func NewFeedbackQuestionText(text string) FeedbackQuestionText {
	return FeedbackQuestionText(text)
}

var (
	ErrFeedbackQuestionTextEmpty   = errors.New("question text is empty")
	ErrFeedbackQuestionTextTooLong = errors.New("question text is too long")
)

func (t FeedbackQuestionText) Validate() error {
	if len(t) == 0 {
		return ErrFeedbackQuestionTextEmpty
	}

	if utf8.RuneCountInString(string(t)) > 256 {
		return ErrFeedbackQuestionTextTooLong
	}

	return nil
}

func NewFeedbackQuestionOrder(order int) FeedbackQuestionOrder {
	return FeedbackQuestionOrder(order)
}

func NewFeedbackComment(comment string) FeedbackComment {
	return FeedbackComment(comment)
}

var (
	ErrFeedbackCommentTooLong = errors.New("feedback comment is too long")
)

func (c FeedbackComment) Validate() error {
	if utf8.RuneCountInString(string(c)) > 2000 {
		return ErrFeedbackCommentTooLong
	}

	return nil
}
