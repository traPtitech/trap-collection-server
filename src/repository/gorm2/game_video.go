package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	gameVideoTypeMp4 = "mp4"
)

type GameVideo struct {
	db *DB
}

var videoTypeSetupGroup = &singleflight.Group{}

func NewGameVideo(db *DB) (*GameVideo, error) {
	ctx := context.Background()

	gormDB, err := db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	/*
		実際の運用では並列で実行されないが、
		テストで並列に実行されるため、
		singleflightを使っている
	*/
	_, err, _ = videoTypeSetupGroup.Do("setupVideoTypeTable", func() (interface{}, error) {
		return nil, setupVideoTypeTable(gormDB)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to setup video type table: %w", err)
	}

	return &GameVideo{
		db: db,
	}, nil
}

func setupVideoTypeTable(db *gorm.DB) error {
	videoTypes := []GameVideoTypeTable{
		{
			Name:   gameVideoTypeMp4,
			Active: true,
		},
	}

	for _, videoType := range videoTypes {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", videoType.Name).
			FirstOrCreate(&videoType).Error
		if err != nil {
			return fmt.Errorf("failed to create role type: %w", err)
		}
	}

	return nil
}

func (gv *GameVideo) SaveGameVideo(ctx context.Context, gameID values.GameID, video *domain.GameVideo) error {
	gormDB, err := gv.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var videoTypeName string
	switch video.GetType() {
	case values.GameVideoTypeMp4:
		videoTypeName = gameVideoTypeMp4
	default:
		return fmt.Errorf("invalid video type: %d", video.GetType())
	}

	var videoType GameVideoTypeTable
	err = gormDB.
		Where("name = ?", videoTypeName).
		Select("id").
		First(&videoType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	videoTypeID := videoType.ID

	err = gormDB.Create(&GameVideoTable{
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

func (gv *GameVideo) GetLatestGameVideo(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.GameVideo, error) {
	gormDB, err := gv.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	gormDB, err = gv.db.setLock(gormDB, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var video GameVideoTable
	err = gormDB.
		Joins("GameVideoType").
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at DESC").
		First(&video).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game video: %w", err)
	}

	var videoType values.GameVideoType
	switch video.GameVideoType.Name {
	case gameVideoTypeMp4:
		videoType = values.GameVideoTypeMp4
	default:
		return nil, fmt.Errorf("invalid video type: %s", video.GameVideoType.Name)
	}

	return domain.NewGameVideo(
		values.NewGameVideoIDFromUUID(video.ID),
		videoType,
		video.CreatedAt,
	), nil
}
