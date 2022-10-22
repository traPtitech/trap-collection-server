package gorm2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
)

var _ repository.ProductKey = (*ProductKey)(nil)

type ProductKey struct {
	db *DB
}

func NewProductKey(db *DB) *ProductKey {
	return &ProductKey{
		db: db,
	}
}

func (productKey *ProductKey) SaveProductKeys(ctx context.Context, editionID values.LauncherVersionID, productKeys []*domain.LauncherUser) error {
	if len(productKeys) == 0 {
		return nil
	}

	db, err := productKey.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var productKeyStatus []migrate.ProductKeyStatusTable2
	err = db.
		Session(&gorm.Session{}).
		Where("active = ?", true).
		Find(&productKeyStatus).Error
	if err != nil {
		return fmt.Errorf("failed to get product key status: %w", err)
	}

	statusMap := make(map[values.LauncherUserStatus]int)
	for _, status := range productKeyStatus {
		switch status.Name {
		case migrate.ProductKeyStatusInactive:
			statusMap[values.LauncherUserStatusInactive] = status.ID
		case migrate.ProductKeyStatusActive:
			statusMap[values.LauncherUserStatusActive] = status.ID
		default:
			log.Printf("invalid product key status: %s\n", status.Name)
		}
	}

	dbProductKeys := make([]*migrate.ProductKeyTable2, 0, len(productKeys))
	for _, key := range productKeys {
		statusID, ok := statusMap[key.GetStatus()]
		if !ok {
			return fmt.Errorf("invalid product key status: %d", key.GetStatus())
		}

		dbProductKeys = append(dbProductKeys, &migrate.ProductKeyTable2{
			ID:         uuid.UUID(key.GetID()),
			EditionID:  uuid.UUID(editionID),
			StatusID:   statusID,
			ProductKey: string(key.GetProductKey()),
			CreatedAt:  time.Now(),
		})
	}

	err = db.Create(&dbProductKeys).Error
	if err != nil {
		return fmt.Errorf("failed to create product keys: %w", err)
	}

	return nil
}

func (productKey *ProductKey) UpdateProductKey(ctx context.Context, key *domain.LauncherUser) error {
	db, err := productKey.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var dbStatus string
	switch key.GetStatus() {
	case values.LauncherUserStatusInactive:
		dbStatus = migrate.ProductKeyStatusInactive
	case values.LauncherUserStatusActive:
		dbStatus = migrate.ProductKeyStatusActive
	default:
		return fmt.Errorf("invalid product key status: %d", key.GetStatus())
	}

	var productKeyStatus migrate.ProductKeyStatusTable2
	err = db.
		Session(&gorm.Session{}).
		Where("name = ?", dbStatus).
		Take(&productKeyStatus).Error
	if err != nil {
		return fmt.Errorf("failed to get product key: %w", err)
	}

	result := db.
		Where("id = ?", uuid.UUID(key.GetID())).
		Updates(&migrate.ProductKeyTable2{
			ProductKey: string(key.GetProductKey()),
			StatusID:   productKeyStatus.ID,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update product key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordUpdated
	}

	return nil
}

func (productKey *ProductKey) GetProductKeys(ctx context.Context, editionID values.LauncherVersionID, statuses []values.LauncherUserStatus, lockType repository.LockType) ([]*domain.LauncherUser, error) {
	db, err := productKey.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = productKey.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var dbStatuses []string
	for _, status := range statuses {
		switch status {
		case values.LauncherUserStatusInactive:
			dbStatuses = append(dbStatuses, migrate.ProductKeyStatusInactive)
		case values.LauncherUserStatusActive:
			dbStatuses = append(dbStatuses, migrate.ProductKeyStatusActive)
		default:
			return nil, fmt.Errorf("invalid product key status: %d", status)
		}
	}

	var dbProductKeys []migrate.ProductKeyTable2
	err = db.
		Joins("Status").
		Where("edition_id = ?", uuid.UUID(editionID)).
		Where("Status.name IN ?", dbStatuses).
		Find(&dbProductKeys).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get product keys: %w", err)
	}

	productKeys := make([]*domain.LauncherUser, 0, len(dbProductKeys))
	for _, dbProductKey := range dbProductKeys {
		var status values.LauncherUserStatus
		switch dbProductKey.Status.Name {
		case migrate.ProductKeyStatusInactive:
			status = values.LauncherUserStatusInactive
		case migrate.ProductKeyStatusActive:
			status = values.LauncherUserStatusActive
		default:
			log.Printf("invalid product key status: %s\n", dbProductKey.Status.Name)
			continue
		}
		keyValue := domain.NewProductKey(
			values.NewLauncherUserIDFromUUID(dbProductKey.ID),
			values.NewLauncherUserProductKeyFromString(dbProductKey.ProductKey),
			status,
			dbProductKey.CreatedAt,
		)

		productKeys = append(productKeys, keyValue)
	}

	return productKeys, nil
}

func (productKey *ProductKey) GetProductKey(ctx context.Context, productKeyID values.LauncherUserID, lockType repository.LockType) (*domain.LauncherUser, error) {
	db, err := productKey.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = productKey.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var dbProductKey migrate.ProductKeyTable2
	err = db.
		Joins("Status").
		Where("product_keys.id = ?", uuid.UUID(productKeyID)).
		Take(&dbProductKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product key: %w", err)
	}

	var status values.LauncherUserStatus
	switch dbProductKey.Status.Name {
	case migrate.ProductKeyStatusInactive:
		status = values.LauncherUserStatusInactive
	case migrate.ProductKeyStatusActive:
		status = values.LauncherUserStatusActive
	default:
		log.Printf("invalid product key status: %s\n", dbProductKey.Status.Name)
	}

	keyValue := domain.NewProductKey(
		values.NewLauncherUserIDFromUUID(dbProductKey.ID),
		values.NewLauncherUserProductKeyFromString(dbProductKey.ProductKey),
		status,
		dbProductKey.CreatedAt,
	)

	return keyValue, nil
}

func (productKey *ProductKey) GetProductKeyByKey(ctx context.Context, productKeyID values.LauncherUserProductKey) (*domain.LauncherUser, error) {
	db, err := productKey.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = productKey.db.setLock(db, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var dbProductKey migrate.ProductKeyTable2
	err = db.
		Joins("Status").
		Where("product_key = ?", string(productKeyID)).
		Take(&dbProductKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher user: %w", err)
	}

	var status values.LauncherUserStatus
	switch dbProductKey.Status.Name {
	case migrate.ProductKeyStatusInactive:
		status = values.LauncherUserStatusInactive
	case migrate.ProductKeyStatusActive:
		status = values.LauncherUserStatusActive
	default:
		log.Printf("invalid product key status: %s\n", dbProductKey.Status.Name)
	}
	keyValue := domain.NewProductKey(
		values.NewLauncherUserIDFromUUID(dbProductKey.ID),
		values.NewLauncherUserProductKeyFromString(dbProductKey.ProductKey),
		status,
		dbProductKey.CreatedAt,
	)

	return keyValue, nil
}
