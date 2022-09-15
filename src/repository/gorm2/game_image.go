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

const (
	gameImageTypeJpeg = "jpeg"
	gameImageTypePng  = "png"
	gameImageTypeGif  = "gif"
)

type GameImage struct {
	db *DB
}

func NewGameImage(db *DB) *GameImage {
	return &GameImage{
		db: db,
	}
}

func (gi *GameImage) SaveGameImage(ctx context.Context, gameID values.GameID, image *domain.GameImage) error {
	gormDB, err := gi.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var imageTypeName string
	switch image.GetType() {
	case values.GameImageTypeJpeg:
		imageTypeName = gameImageTypeJpeg
	case values.GameImageTypePng:
		imageTypeName = gameImageTypePng
	case values.GameImageTypeGif:
		imageTypeName = gameImageTypeGif
	default:
		return fmt.Errorf("invalid image type: %d", image.GetType())
	}

	var imageType migrate.GameImageTypeTable
	err = gormDB.
		Where("name = ?", imageTypeName).
		Select("id").
		First(&imageType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	imageTypeID := imageType.ID

	err = gormDB.Create(&migrate.GameImageTable{
		ID:          uuid.UUID(image.GetID()),
		GameID:      uuid.UUID(gameID),
		ImageTypeID: imageTypeID,
		CreatedAt:   image.GetCreatedAt(),
	}).Error
	if err != nil {
		return fmt.Errorf("failed to create game image: %w", err)
	}

	return nil
}

func (gi *GameImage) GetLatestGameImage(ctx context.Context, gameID values.GameID, lockType repository.LockType) (*domain.GameImage, error) {
	gormDB, err := gi.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	gormDB, err = gi.db.setLock(gormDB, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var image migrate.GameImageTable
	err = gormDB.
		Joins("GameImageType").
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at DESC").
		First(&image).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game image: %w", err)
	}

	var imageType values.GameImageType
	switch image.GameImageType.Name {
	case gameImageTypeJpeg:
		imageType = values.GameImageTypeJpeg
	case gameImageTypePng:
		imageType = values.GameImageTypePng
	case gameImageTypeGif:
		imageType = values.GameImageTypeGif
	default:
		return nil, fmt.Errorf("invalid image type: %s", image.GameImageType.Name)
	}

	return domain.NewGameImage(
		values.GameImageIDFromUUID(image.ID),
		imageType,
		image.CreatedAt,
	), nil
}
