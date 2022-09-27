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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

type LauncherUser struct {
	db *DB
}

func NewLauncherUser(db *DB) *LauncherUser {
	return &LauncherUser{
		db: db,
	}
}

func (lu *LauncherUser) CreateLauncherUsers(ctx context.Context, launcherVersionID values.LauncherVersionID, launcherUsers []*domain.LauncherUser) ([]*domain.LauncherUser, error) {
	if len(launcherUsers) == 0 {
		return []*domain.LauncherUser{}, nil
	}

	db, err := lu.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	dbLauncherUsers := make([]*migrate.LauncherUserTable, 0, len(launcherUsers))
	for _, launcherUser := range launcherUsers {
		dbLauncherUsers = append(dbLauncherUsers, &migrate.LauncherUserTable{
			ID:                uuid.UUID(launcherUser.GetID()),
			LauncherVersionID: uuid.UUID(launcherVersionID),
			ProductKey:        string(launcherUser.GetProductKey()),
			CreatedAt:         time.Now(),
		})
	}

	err = db.Create(&dbLauncherUsers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher users: %w", err)
	}

	return launcherUsers, nil
}

func (lu *LauncherUser) DeleteLauncherUser(ctx context.Context, launcherUserID values.LauncherUserID) error {
	db, err := lu.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := db.Delete(&migrate.LauncherUserTable{ID: uuid.UUID(launcherUserID)})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to delete launcher user: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}

	return nil
}

func (lu *LauncherUser) GetLauncherUserByProductKey(ctx context.Context, productKey values.LauncherUserProductKey) (*domain.LauncherUser, error) {
	db, err := lu.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var dbLauncherUser migrate.LauncherUserTable
	err = db.
		Where("product_key = ?", string(productKey)).
		Take(&dbLauncherUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher user: %w", err)
	}

	return domain.NewLauncherUser(
		values.NewLauncherUserIDFromUUID(dbLauncherUser.ID),
		values.NewLauncherUserProductKeyFromString(dbLauncherUser.ProductKey),
	), nil
}
