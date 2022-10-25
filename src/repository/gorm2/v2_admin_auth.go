package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
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
