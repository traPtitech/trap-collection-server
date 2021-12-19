package gorm2

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

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

func (lv *LauncherVersion) GetLauncherUsersByLauncherVersionID(ctx context.Context, launcherVersionID values.LauncherVersionID) ([]*domain.LauncherUser, error) {
	db, err := lv.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var dbLauncherUsers []*LauncherUserTable
	err = db.
		Where("launcher_version_id = ?", uuid.UUID(launcherVersionID)).
		Find(&dbLauncherUsers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher users: %w", err)
	}

	launcherUsers := make([]*domain.LauncherUser, 0, len(dbLauncherUsers))
	for _, dbLauncherUser := range dbLauncherUsers {
		launcherUsers = append(launcherUsers, domain.NewLauncherUser(
			values.NewLauncherUserIDFromUUID(dbLauncherUser.ID),
			values.NewLauncherUserProductKeyFromString(dbLauncherUser.ProductKey),
		))
	}

	return launcherUsers, nil
}

func (lv *LauncherVersion) GetLauncherVersionAndUserAndSessionByAccessToken(ctx context.Context, accessToken values.LauncherSessionAccessToken) (*domain.LauncherVersion, *domain.LauncherUser, *domain.LauncherSession, error) {
	db, err := lv.db.getDB(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get db: %w", err)
	}

	type LauncherVersion struct {
		// LancherVersionとGameのmany2manyを追加したら重複Fieldがないのに重複Fieldがあるというエラーが出たので、暫定対処
		LauncherVersion struct {
			ID               uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
			Name             string         `gorm:"type:varchar(32);not null;unique"`
			QuestionnaireURL sql.NullString `gorm:"type:text;default:NULL"`
			CreatedAt        time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
			DeletedAt        gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
		} `gorm:"embedded;embeddedPrefix:launcher_versions_"`
		LauncherUser    LauncherUserTable    `gorm:"embedded;embeddedPrefix:launcher_users_"`
		LauncherSession LauncherSessionTable `gorm:"embedded;embeddedPrefix:launcher_sessions_"`
	}

	var scanStruct LauncherVersion
	err = db.
		Unscoped(). //TakeでJOIN結果を取るため、Unscopedをしつつ自前でdeleted_at IS NULLを指定している
		Table("launcher_versions").
		Where("launcher_versions.deleted_at IS NULL").
		Joins("INNER JOIN launcher_users ON launcher_versions.id = launcher_users.launcher_version_id AND launcher_users.deleted_at IS NULL").
		Joins("INNER JOIN launcher_sessions ON launcher_users.id = launcher_sessions.launcher_user_id AND launcher_sessions.deleted_at IS NULL").
		Where("launcher_sessions.access_token = ?", accessToken).
		Select("launcher_versions.id AS launcher_versions_id, launcher_versions.name AS launcher_versions_name, launcher_versions.questionnaire_url AS launcher_versions_questionnaire_url, launcher_versions.created_at AS launcher_versions_created_at, " +
			"launcher_users.id AS launcher_users_id, launcher_users.product_key AS launcher_users_product_key, " +
			"launcher_sessions.id AS launcher_sessions_id, launcher_sessions.access_token AS launcher_sessions_access_token, launcher_sessions.expires_at AS launcher_sessions_expires_at").
		Take(&scanStruct).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	var launcherVersion *domain.LauncherVersion
	if scanStruct.LauncherVersion.QuestionnaireURL.Valid {
		questionnaireURL, err := url.Parse(scanStruct.LauncherVersion.QuestionnaireURL.String)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse questionnaire url: %w", err)
		}

		launcherVersion = domain.NewLauncherVersionWithQuestionnaire(
			values.NewLauncherVersionIDFromUUID(scanStruct.LauncherVersion.ID),
			values.NewLauncherVersionName(scanStruct.LauncherVersion.Name),
			values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
			scanStruct.LauncherVersion.CreatedAt,
		)
	} else {
		launcherVersion = domain.NewLauncherVersionWithoutQuestionnaire(
			values.NewLauncherVersionIDFromUUID(scanStruct.LauncherVersion.ID),
			values.NewLauncherVersionName(scanStruct.LauncherVersion.Name),
			scanStruct.LauncherVersion.CreatedAt,
		)
	}

	launcherUser := domain.NewLauncherUser(
		values.NewLauncherUserIDFromUUID(scanStruct.LauncherUser.ID),
		values.NewLauncherUserProductKeyFromString(scanStruct.LauncherUser.ProductKey),
	)

	launcherSession := domain.NewLauncherSession(
		values.NewLauncherSessionIDFromUUID(scanStruct.LauncherSession.ID),
		values.NewLauncherSessionAccessTokenFromString(scanStruct.LauncherSession.AccessToken),
		scanStruct.LauncherSession.ExpiresAt,
	)

	return launcherVersion, launcherUser, launcherSession, nil
}
