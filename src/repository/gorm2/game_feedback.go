package gorm2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

type GameFeedback struct {
	db *DB
}

func (g *GameFeedback) GetFeedbackQuestions(ctx context.Context, gameID values.GameID, lockType repository.LockType) ([]*domain.FeedbackQuestion, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}
	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}
	var tables []schema.GameFeedbackQuestionTable
	err = db.
		Where("game_id = ? AND archived_at IS NULL", uuid.UUID(gameID)).
		Order("question_order ASC").
		Find(&tables).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback questions: %w", err)
	}
	questions := make([]*domain.FeedbackQuestion, 0, len(tables))
	for _, table := range tables {
		var archivedAt *time.Time
		if table.ArchivedAt.Valid {
			archivedAt = &table.ArchivedAt.Time
		}
		questions = append(questions, domain.NewFeedbackQuestion(
			values.NewFeedbackQuestionIDFromUUID(table.ID),
			values.NewGameIDFromUUID(table.GameID),
			values.NewFeedbackQuestionText(table.QuestionText),
			values.FeedbackAnswerType(table.AnswerType),
			values.NewFeedbackQuestionOrder(table.QuestionOrder),
			table.CreatedAt,
			archivedAt,
		))
	}
	return questions, nil
}

var _ repository.GameFeedback = (*GameFeedback)(nil)

func NewGameFeedback(db *DB) *GameFeedback {
	return &GameFeedback{
		db: db,
	}
}

func (g *GameFeedback) GetFeedbackConfig(ctx context.Context, gameID values.GameID, lockType repository.LockType) (bool, error) {
	db, err := g.db.getDB(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return false, fmt.Errorf("failed to set lock: %w", err)
	}

	var config schema.GameFeedbackConfigTable
	err = db.Where("game_id = ?", uuid.UUID(gameID)).Take(&config).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, repository.ErrRecordNotFound
	}
	if err != nil {
		return false, fmt.Errorf("failed to get game feedback config: %w", err)
	}

	return config.Enabled, nil
}
