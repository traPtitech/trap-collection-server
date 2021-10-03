package gorm2

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"gorm.io/gorm"
)

type LauncherVersion struct {
	db *DB
}

func NewLauncherVersion(db *DB) *LauncherVersion {
	return &LauncherVersion{
		db: db,
	}
}

func (lv *LauncherVersion) GetLauncherVersion(ctx context.Context, launcherVersionID values.LauncherVersionID) (*domain.LauncherVersion, error) {
	db, err := lv.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var dbLauncherVersion LauncherVersionTable
	err = db.
		Where("id = ?", uuid.UUID(launcherVersionID)).
		Take(&dbLauncherVersion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	var launcherVersion *domain.LauncherVersion
	if dbLauncherVersion.QuestionnaireURL.Valid {
		questionnaireURL, err := url.Parse(dbLauncherVersion.QuestionnaireURL.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse questionnaire url: %w", err)
		}

		launcherVersion = domain.NewLauncherVersionWithQuestionnaire(
			values.NewLauncherVersionIDFromUUID(dbLauncherVersion.ID),
			values.NewLauncherVersionName(dbLauncherVersion.Name),
			values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
			dbLauncherVersion.CreatedAt,
		)
	} else {
		launcherVersion = domain.NewLauncherVersionWithoutQuestionnaire(
			values.NewLauncherVersionIDFromUUID(dbLauncherVersion.ID),
			values.NewLauncherVersionName(dbLauncherVersion.Name),
			dbLauncherVersion.CreatedAt,
		)
	}

	return launcherVersion, nil
}
