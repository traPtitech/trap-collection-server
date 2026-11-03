package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
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

func (g *GameFeedback) CreateGameFeedback(ctx context.Context, feedback *domain.GameFeedback, answers []*domain.GameFeedbackAnswer) error {
	if feedback == nil {
		return errors.New("game feedback is nil")
	}

	feedbackID := feedback.GetID()
	answerTables := make([]schema.GameFeedbackAnswerTable, 0, len(answers))
	for _, answer := range answers {
		if answer == nil {
			return errors.New("game feedback answer is nil")
		}
		if answer.GetFeedbackID() != feedbackID {
			return fmt.Errorf("game feedback answer %s belongs to feedback %s, not %s", answer.GetID(), answer.GetFeedbackID(), feedbackID)
		}
		answerTable := schema.GameFeedbackAnswerTable{
			ID:         answer.GetID().UUID(),
			FeedbackID: answer.GetFeedbackID().UUID(),
			QuestionID: answer.GetQuestionID().UUID(),
			Answer:     answer.GetAnswer(),
		}
		answerTables = append(answerTables, answerTable)
	}

	var comment sql.NullString
	if feedback.GetComment() != nil {
		comment = sql.NullString{
			String: string(*feedback.GetComment()),
			Valid:  true,
		}
	}

	feedbackTable := schema.GameFeedbackTable{
		ID:            feedbackID.UUID(),
		GameVersionID: uuid.UUID(feedback.GetGameVersionID()),
		Comment:       comment,
		CreatedAt:     feedback.GetCreatedAt(),
	}

	db, err := g.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("get db: %w", err)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		transactionDB := tx.WithContext(ctx)
		if err := transactionDB.Create(&feedbackTable).Error; err != nil {
			return fmt.Errorf("create game feedback: %w", err)
		}
		if len(answerTables) == 0 {
			return nil
		}
		if err := transactionDB.Create(&answerTables).Error; err != nil {
			return fmt.Errorf("create game feedback answers: %w", err)
		}
		return nil
	})
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return repository.ErrDuplicatedUniqueKey
		case 1452:
			return repository.ErrForeignKeyViolated
		}
	}
	return fmt.Errorf("create game feedback: %w", err)
}
