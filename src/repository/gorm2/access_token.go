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

	var edition migrate.EditionTable2
	err = db.
		Where("editions.deleted_at IS NULL").
		Joins("INNER JOIN product_keys ON product_keys.edition_id = editions.id AND product_keys.deleted_at IS NULL").
		Joins("INNER JOIN product_key_statuses ON product_key_statuses.id = product_keys.status_id AND product_key_statuses.active").
		Joins("INNER JOIN access_tokens ON access_tokens.product_key_id = product_keys.id AND access_tokens.deleted_at IS NULL").
		Where("access_tokens.access_token = ?", string(token)).
		Take(&edition).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get edition: %w", err)
	}

	if len(edition.ProductKeys) == 0 {
		return nil, repository.ErrRecordNotFound
	}
	dbProductKey := edition.ProductKeys[0]

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

	if len(dbProductKey.AccessTokens) == 0 {
		return nil, repository.ErrRecordNotFound
	}
	dbAccessToken := dbProductKey.AccessTokens[0]

	var domainEdition *domain.LauncherVersion
	if edition.QuestionnaireURL.Valid {
		questionnaireURL, err := url.Parse(edition.QuestionnaireURL.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse questionnaire url: %w", err)
		}

		domainEdition = domain.NewLauncherVersionWithQuestionnaire(
			values.NewLauncherVersionIDFromUUID(edition.ID),
			values.NewLauncherVersionName(edition.Name),
			values.NewLauncherVersionQuestionnaireURL(questionnaireURL),
			edition.CreatedAt,
		)
	} else {
		domainEdition = domain.NewLauncherVersionWithoutQuestionnaire(
			values.NewLauncherVersionIDFromUUID(edition.ID),
			values.NewLauncherVersionName(edition.Name),
			edition.CreatedAt,
		)
	}

	return &repository.AccessTokenInfo{
		AccessToken: domain.NewLauncherSession(
			values.NewLauncherSessionIDFromUUID(dbAccessToken.ID),
			values.NewLauncherSessionAccessTokenFromString(dbAccessToken.AccessToken),
			dbAccessToken.ExpiresAt,
		),
		ProductKey: key,
		Edition:    domainEdition,
	}, nil
}
