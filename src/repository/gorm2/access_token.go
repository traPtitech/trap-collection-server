package gorm2

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

var _ repository.AccessToken = (*AccessToken)(nil)

type AccessToken struct {
	db *DB
}

func NewAccessToken(db *DB) *AccessToken {
	return &AccessToken{
		db: db,
	}
}

func (accessToken *AccessToken) SaveAccessToken(ctx context.Context, productKeyID values.LauncherUserID, token *domain.LauncherSession) error {
	db, err := accessToken.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	err = db.
		Create(&migrate.AccessTokenTable2{
			ID:           uuid.UUID(token.GetID()),
			ProductKeyID: uuid.UUID(productKeyID),
			AccessToken:  string(token.GetAccessToken()),
			ExpiresAt:    token.GetExpiresAt(),
			CreatedAt:    time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to create access token: %w", err)
	}

	return nil
}

func (accessToken *AccessToken) GetAccessTokenInfo(ctx context.Context, token values.LauncherSessionAccessToken, lockType repository.LockType) (*repository.AccessTokenInfo, error) {
	db, err := accessToken.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = accessToken.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}
	type Edition struct {
		// EditionとGameのmany2manyを追加したら重複Fieldがないのに重複Fieldがあるというエラーが出たので、暫定対処
		Edition struct {
			ID               uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
			Name             string         `gorm:"type:varchar(32);not null;unique"`
			QuestionnaireURL sql.NullString `gorm:"type:text;default:NULL"`
			CreatedAt        time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
			DeletedAt        gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
		} `gorm:"embedded;embeddedPrefix:edition_"`
		ProductKey       migrate.ProductKeyTable2       `gorm:"embedded;embeddedPrefix:product_key_"`
		ProductKeyStatus migrate.ProductKeyStatusTable2 `gorm:"embedded;embeddedPrefix:product_key_status_"`
		AccessToken      migrate.AccessTokenTable2      `gorm:"embedded;embeddedPrefix:access_token_"`
	}
	var scanStruct Edition

	selectMaps := []string{
		"editions.id AS edition_id",
		"editions.name AS edition_name",
		"editions.questionnaire_url AS edition_questionnaire_url",
		"editions.created_at AS edition_created_at",
		"product_keys.id AS product_key_id",
		"product_keys.product_key AS product_key_product_key",
		"product_keys.created_at AS product_key_created_at",
		"product_key_statuses.name AS product_key_status_name",
		"access_tokens.id AS access_token_id",
		"access_tokens.access_token AS access_token_access_token",
		"access_tokens.expires_at AS access_token_expires_at",
		"access_tokens.created_at AS access_token_created_at",
	}

	err = db.
		Unscoped().
		Table("editions").
		Where("editions.deleted_at IS NULL").
		Joins("INNER JOIN product_keys ON product_keys.edition_id = editions.id").
		Joins("INNER JOIN product_key_statuses ON product_key_statuses.id = product_keys.status_id AND product_key_statuses.active").
		Joins("INNER JOIN access_tokens ON access_tokens.product_key_id = product_keys.id AND access_tokens.deleted_at IS NULL").
		Where("access_tokens.access_token = ?", string(token)).
		Select(selectMaps).
		Take(&scanStruct).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition: %w", err)
	}

	dbEdition := scanStruct.Edition
	dbProductKey := scanStruct.ProductKey
	dbAccessToken := scanStruct.AccessToken

	var status values.LauncherUserStatus
	switch dbProductKey.Status.Name {
	case migrate.ProductKeyStatusInactive:
		status = values.LauncherUserStatusInactive
	case migrate.ProductKeyStatusActive:
		status = values.LauncherUserStatusActive
	}
	key := domain.NewProductKey(
		values.NewLauncherUserIDFromUUID(dbProductKey.ID),
		values.NewLauncherUserProductKeyFromString(dbProductKey.ProductKey),
		status,
		dbProductKey.CreatedAt,
	)

	var edition *domain.Edition
	if dbEdition.QuestionnaireURL.Valid {
		questionnaireURL, err := url.Parse(dbEdition.QuestionnaireURL.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse questionnaire url: %w", err)
		}

		edition = domain.NewEditionWithQuestionnaire(
			values.NewEditionIDFromUUID(dbEdition.ID),
			values.NewEditionName(dbEdition.Name),
			values.NewEditionQuestionnaireURL(questionnaireURL),
			dbEdition.CreatedAt,
		)
	} else {
		edition = domain.NewEditionWithoutQuestionnaire(
			values.NewEditionIDFromUUID(dbEdition.ID),
			values.NewEditionName(dbEdition.Name),
			dbEdition.CreatedAt,
		)
	}

	return &repository.AccessTokenInfo{
		AccessToken: domain.NewLauncherSession(
			values.NewLauncherSessionIDFromUUID(dbAccessToken.ID),
			values.NewLauncherSessionAccessTokenFromString(dbAccessToken.AccessToken),
			dbAccessToken.ExpiresAt,
		),
		ProductKey: key,
		Edition:    edition,
	}, nil
}
