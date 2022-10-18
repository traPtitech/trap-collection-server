package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
	"gorm.io/gorm"
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

func (gameVideo *GameVideoV2) GetGameVideo(ctx context.Context, gameVideoID values.GameVideoID, lockType repository.LockType) (*repository.GameVideoInfo, error) {
	db, err := gameVideo.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameVideo.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var video migrate.GameVideoTable2
	err = db.
		Joins("GameVideoType").
		Where("v2_game_video.id = ?", uuid.UUID(gameVideoID)).
		Take(&video).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game video: %w", err)
	}

	var videoType values.GameVideoType
	if video.GameVideoType.Name == migrate.GameVideoTypeMp4 {
		videoType = values.GameVideoTypeMp4
	} else {
		return nil, fmt.Errorf("invalid video type: %s", video.GameVideoType.Name)
	}

	return &repository.GameVideoInfo{
		GameVideo: domain.NewGameVideo(
			values.NewGameVideoIDFromUUID(video.ID),
			videoType,
			video.CreatedAt,
		),
		GameID: values.NewGameIDFromUUID(video.GameID),
	}, nil
}
