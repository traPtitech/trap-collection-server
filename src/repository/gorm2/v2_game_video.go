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
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

type GameVideoV2 struct {
	db *DB
}

func convertGameVideoType(migrateGameVideoType string) (values.GameVideoType, error) {
	switch migrateGameVideoType {
	case migrate.GameVideoTypeMp4:
		return values.GameVideoTypeMp4, nil
	case migrate.GameVideoTypeM4v:
		return values.GameVideoTypeM4v, nil
	case migrate.GameVideoTypeMkv:
		return values.GameVideoTypeMkv, nil
	default:
		return 0, fmt.Errorf("invalid video type: %s", migrateGameVideoType)
	}
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
	switch video.GetType() {
	case values.GameVideoTypeMp4:
		videoTypeName = migrate.GameVideoTypeMp4
	case values.GameVideoTypeM4v:
		videoTypeName = migrate.GameVideoTypeM4v
	case values.GameVideoTypeMkv:
		videoTypeName = migrate.GameVideoTypeMkv
	default:
		return fmt.Errorf("invalid video type: %d", video.GetType())
	}

	var videoType schema.GameVideoTypeTable
	err = db.
		Where("name = ?", videoTypeName).
		Where("active = ?", true).
		Select("id").
		Take(&videoType).Error
	if err != nil {
		return fmt.Errorf("failed to get video type: %w", err)
	}
	videoTypeID := videoType.ID

	err = db.
		Create(&schema.GameVideoTable2{
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

	var video schema.GameVideoTable2
	err = db.
		Joins("GameVideoType").
		Where("v2_game_videos.id = ?", uuid.UUID(gameVideoID)).
		Take(&video).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game video: %w", err)
	}

	videoType, err := convertGameVideoType(video.GameVideoType.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to convert video type: %w", err)
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

func (gameVideo *GameVideoV2) GetGameVideos(ctx context.Context, gameID values.GameID, lockType repository.LockType) ([]*domain.GameVideo, error) {
	db, err := gameVideo.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameVideo.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var videos []schema.GameVideoTable2
	err = db.
		Joins("GameVideoType").
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at DESC").
		Find(&videos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game videos: %w", err)
	}

	gameVideos := make([]*domain.GameVideo, 0, len(videos))
	for _, video := range videos {
		videoType, err := convertGameVideoType(video.GameVideoType.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to convert video type: %w", err)
		}

		gameVideos = append(gameVideos, domain.NewGameVideo(
			values.NewGameVideoIDFromUUID(video.ID),
			videoType,
			video.CreatedAt,
		))
	}

	return gameVideos, nil
}
