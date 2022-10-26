package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

type AdminAuth struct {
	db *DB
}

func NewAdminAuth(db *DB) *AdminAuth {
	return &AdminAuth{
		db: db,
	}
}

func (aa *AdminAuth) AddAdmin(ctx context.Context, userID values.TraPMemberID) error {
	db, err := aa.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	adminTable := migrate.AdminTable{
		UserID: uuid.UUID(userID),
	}

	err = db.Create(&adminTable).Error
	if err != nil {
		return fmt.Errorf("failed to add admin: %w", err)
	}
	return nil
}

func (aa *AdminAuth) GetAdmins(ctx context.Context) ([]values.TraPMemberID, error) {
	db, err := aa.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	var admins []migrate.AdminTable
	err = db.Find(&admins).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}

	adminsID := make([]values.TraPMemberID, 0)
	for _, admin := range admins {
		adminsID = append(adminsID, values.NewTrapMemberID(admin.UserID))
	}
	return adminsID, nil
}

func (aa *AdminAuth) DeleteAdmin(ctx context.Context, userID values.TraPMemberID) error {
	db, err := aa.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	result := db.
		Where("user_id = ?", userID).
		Delete(&migrate.AdminTable{})
	err = result.Error
	if err != nil {
		return fmt.Errorf("failed to remove admin: %w", err)
	}

	if result.RowsAffected == 0 {
		return repository.ErrNoRecordDeleted
	}
	return nil
}
