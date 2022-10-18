package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
)

type GameVideoV2 struct {
	db *DB
}

func NewGameVideoV2(db *DB) *GameVideoV2 {
	return &GameVideoV2{
		db: db,
	}
}

func (gameVideo *GameVideoV2) SaveGameVideo(ctx context.Context, gameID values.GameID, video *domain.GameVideo) error {
	db, err := gameVideo.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var videoTypeName string
	if video.GetType() == values.GameVideoTypeMp4 {
		videoTypeName = migrate.GameVideoTypeMp4
	} else {
		return fmt.Errorf("invalid video type: %d", video.GetType())
	}

	var videoType migrate.GameVideoTypeTable
	err = db.
		Where("name = ?", videoTypeName).
		Select("id").
		Take(&videoType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	videoTypeID := videoType.ID

	err = db.
		Create(&migrate.GameVideoTable2{
			ID:          uuid.UUID(video.GetID()),
			GameID:      uuid.UUID(gameID),
			VideoTypeID: videoTypeID,
			CreatedAt:   video.GetCreatedAt(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to create game video: %w", err)
	}

	return nil
}
