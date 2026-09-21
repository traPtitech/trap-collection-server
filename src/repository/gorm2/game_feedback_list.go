package gorm2

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

func (g *GameFeedback) GetFeedbackQuestionsIncludingArchived(ctx context.Context, gameID values.GameID, lockType repository.LockType) ([]*domain.FeedbackQuestion, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var questions []schema.GameFeedbackQuestionTable
	err = db.
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("question_order").
		Order("id").
		Find(&questions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback questions: %w", err)
	}

	result := make([]*domain.FeedbackQuestion, 0, len(questions))
	for _, question := range questions {
		var archivedAt *time.Time
		if question.ArchivedAt.Valid {
			archivedAt = &question.ArchivedAt.Time
		}
		result = append(result, domain.NewFeedbackQuestion(
			values.FeedbackQuestionID(question.ID),
			values.GameID(question.GameID),
			values.NewFeedbackQuestionText(question.QuestionText),
			values.FeedbackAnswerType(question.AnswerType),
			values.NewFeedbackQuestionOrder(question.QuestionOrder),
			question.CreatedAt,
			archivedAt,
		))
	}

	return result, nil
}

func (gf *GameFeedback) GetGameFeedbacksByGameID(ctx context.Context, gameID values.GameID, limit, offset int) ([]*repository.GameFeedbackWithAnswers, int, error) {
	db, err := gf.db.getDB(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get db: %w", err)
	}

	query := db.Model(&schema.GameFeedbackTable{}).
		Joins("JOIN v2_game_versions ON v2_game_versions.id = game_feedbacks.game_version_id").
		Where("v2_game_versions.game_id = ?", uuid.UUID(gameID))

	return getGameFeedbacks(query, limit, offset)
}

func getGameFeedbacks(query *gorm.DB, limit, offset int) ([]*repository.GameFeedbackWithAnswers, int, error) {
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count game feedbacks: %w", err)
	}

	var feedbacks []schema.GameFeedbackTable
	err := query.
		Session(&gorm.Session{}).
		Preload("Answers", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN feedback_questions ON feedback_questions.id = game_feedback_answers.question_id AND feedback_questions.deleted_at IS NULL").
				Order("game_feedback_answers.id")
		}).
		Order("game_feedbacks.created_at DESC").
		Order("game_feedbacks.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&feedbacks).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get game feedbacks: %w", err)
	}

	feedbacksWithAnswers := make([]*repository.GameFeedbackWithAnswers, 0, len(feedbacks))
	for _, feedback := range feedbacks {
		feedbacksWithAnswers = append(feedbacksWithAnswers, gameFeedbackWithAnswersFromTable(&feedback))
	}

	return feedbacksWithAnswers, int(total), nil
}

func gameFeedbackWithAnswersFromTable(feedback *schema.GameFeedbackTable) *repository.GameFeedbackWithAnswers {
	var comment *values.FeedbackComment
	if feedback.Comment.Valid {
		value := values.NewFeedbackComment(feedback.Comment.String)
		comment = &value
	}

	answers := make([]*domain.GameFeedbackAnswer, 0, len(feedback.Answers))
	for _, answer := range feedback.Answers {
		answers = append(answers, domain.NewGameFeedbackAnswer(
			values.GameFeedbackAnswerID(answer.ID),
			values.GameFeedbackID(answer.FeedbackID),
			values.FeedbackQuestionID(answer.QuestionID),
			answer.Answer,
		))
	}

	return &repository.GameFeedbackWithAnswers{
		Feedback: domain.NewGameFeedback(
			values.GameFeedbackID(feedback.ID),
			values.GameVersionID(feedback.GameVersionID),
			comment,
			feedback.CreatedAt,
		),
		Answers: answers,
	}
}
