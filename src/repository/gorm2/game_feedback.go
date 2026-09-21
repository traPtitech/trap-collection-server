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

func (g *GameFeedback) CreateFeedbackQuestions(ctx context.Context, questions []*domain.FeedbackQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	tables := make([]schema.GameFeedbackQuestionTable, 0, len(questions))
	for _, question := range questions {
		tables = append(tables, schema.GameFeedbackQuestionTable{
			ID:            question.GetID().UUID(),
			GameID:        uuid.UUID(question.GetGameID()),
			QuestionText:  string(question.GetQuestionText()),
			AnswerType:    int(question.GetAnswerType()),
			QuestionOrder: int(question.GetQuestionOrder()),
			CreatedAt:     question.GetCreatedAt(),
		})
	}

	if err := db.Create(&tables).Error; err != nil {
		return fmt.Errorf("failed to create feedback questions: %w", err)
	}
	return nil
}

func (g *GameFeedback) UpdateFeedbackQuestions(ctx context.Context, questions []*domain.FeedbackQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(questions))
	for _, question := range questions {
		ids = append(ids, question.GetID().UUID())
	}
	var count int64
	if err := db.Model(&schema.GameFeedbackQuestionTable{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check feedback questions: %w", err)
	}
	if count != int64(len(ids)) {
		return repository.ErrNoRecordUpdated
	}

	for _, question := range questions {
		updates := map[string]any{
			"question_text":  string(question.GetQuestionText()),
			"answer_type":    int(question.GetAnswerType()),
			"question_order": int(question.GetQuestionOrder()),
		}
		if err := db.Model(&schema.GameFeedbackQuestionTable{}).Where("id = ?", question.GetID().UUID()).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update feedback question: %w", err)
		}
	}
	return nil
}

func (g *GameFeedback) ArchiveFeedbackQuestions(ctx context.Context, ids []values.FeedbackQuestionID) error {
	if len(ids) == 0 {
		return nil
	}

	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	uuidIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		uuidIDs = append(uuidIDs, id.UUID())
	}
	if err := db.Model(&schema.GameFeedbackQuestionTable{}).Where("id IN ?", uuidIDs).Update("archived_at", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to archive feedback questions: %w", err)
	}
	return nil
}

func (g *GameFeedback) HasFeedbackAnswers(ctx context.Context, questionIDs []values.FeedbackQuestionID, lockType repository.LockType) (bool, error) {
	if len(questionIDs) == 0 {
		return false, nil
	}
	db, err := g.db.getDB(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get db: %w", err)
	}
	db, err = g.db.setLock(db, lockType)
	if err != nil {
		return false, fmt.Errorf("failed to set lock: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(questionIDs))
	for _, id := range questionIDs {
		ids = append(ids, id.UUID())
	}
	var answer schema.GameFeedbackAnswerTable
	err = db.Where("question_id IN ?", ids).Take(&answer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get feedback answers: %w", err)
	}
	return true, nil
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
