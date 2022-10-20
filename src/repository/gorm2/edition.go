package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

type Edition struct {
	db *DB
}

func NewEdition(db *DB) *Edition {
	return &Edition{
		db: db,
	}
}

func (e *Edition) SaveEdition(ctx context.Context, edition *domain.LauncherVersion) error {
	db, err := e.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	questionnaireURL, err := edition.GetQuestionnaireURL()
	if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
		return fmt.Errorf("failed to get questionnaire url: %w", err)
	}

	var strQuestionnaireURL sql.NullString
	if errors.Is(err, domain.ErrNoQuestionnaire) {
		strQuestionnaireURL = sql.NullString{
			Valid: false,
		}
	} else {
		strQuestionnaireURL = sql.NullString{
			String: (*url.URL)(questionnaireURL).String(),
			Valid:  true,
		}
	}

	err = db.
		Create(&migrate.EditionTable2{
			ID:               uuid.UUID(edition.GetID()),
			Name:             string(edition.GetName()),
			QuestionnaireURL: strQuestionnaireURL,
			CreatedAt:        edition.GetCreatedAt(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to save edition: %w", err)
	}

	return nil
}
