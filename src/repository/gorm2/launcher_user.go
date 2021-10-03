package gorm2

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
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

	dbLauncherUsers := make([]*LauncherUserTable, 0, len(launcherUsers))
	for _, launcherUser := range launcherUsers {
		dbLauncherUsers = append(dbLauncherUsers, &LauncherUserTable{
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
