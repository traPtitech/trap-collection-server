package gorm2

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

type LauncherSession struct {
	db *DB
}

func NewLauncherSession(db *DB) *LauncherSession {
	return &LauncherSession{
		db: db,
	}
}

func (ls *LauncherSession) CreateLauncherSession(ctx context.Context, launcherUserID values.LauncherUserID, launcherSession *domain.LauncherSession) (*domain.LauncherSession, error) {
	db, err := ls.db.getDB(ctx)
	if err != nil {
		return nil, err
	}

	dbLauncherSession := migrate.LauncherSessionTable{
		ID:             uuid.UUID(launcherSession.GetID()),
		LauncherUserID: uuid.UUID(launcherUserID),
		AccessToken:    string(launcherSession.GetAccessToken()),
		ExpiresAt:      launcherSession.GetExpiresAt(),
		CreatedAt:      time.Now(),
	}

	err = db.Create(&dbLauncherSession).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher session: %w", err)
	}

	return launcherSession, nil
}
