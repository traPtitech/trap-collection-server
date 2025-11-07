package gorm2

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema"
	"gorm.io/gorm"
)

type GameImageV2 struct {
	db *DB
}

func NewGameImageV2(db *DB) *GameImageV2 {
	return &GameImageV2{
		db: db,
	}
}

func (gameImage *GameImageV2) SaveGameImage(ctx context.Context, gameID values.GameID, image *domain.GameImage) error {
	db, err := gameImage.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var imageTypeName string
	switch image.GetType() {
	case values.GameImageTypeJpeg:
		imageTypeName = schema.GameImageTypeJpeg
	case values.GameImageTypePng:
		imageTypeName = schema.GameImageTypePng
	case values.GameImageTypeGif:
		imageTypeName = schema.GameImageTypeGif
	default:
		return fmt.Errorf("invalid image type: %d", image.GetType())
	}

	var imageType schema.GameImageTypeTable
	err = db.
		Where("name = ?", imageTypeName).
		Select("id").
		Take(&imageType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	imageTypeID := imageType.ID

	err = db.
		Create(&schema.GameImageTable2{
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

func (gameImage *GameImageV2) GetGameImage(ctx context.Context, gameImageID values.GameImageID, lockType repository.LockType) (*repository.GameImageInfo, error) {
	db, err := gameImage.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameImage.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var image schema.GameImageTable2
	err = db.
		Joins("GameImageType").
		Where("v2_game_images.id = ?", uuid.UUID(gameImageID)).
		Take(&image).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game image: %w", err)
	}

	var imageType values.GameImageType
	switch image.GameImageType.Name {
	case schema.GameImageTypeJpeg:
		imageType = values.GameImageTypeJpeg
	case schema.GameImageTypePng:
		imageType = values.GameImageTypePng
	case schema.GameImageTypeGif:
		imageType = values.GameImageTypeGif
	default:
		return nil, fmt.Errorf("invalid image type: %s", image.GameImageType.Name)
	}

	return &repository.GameImageInfo{
		GameImage: domain.NewGameImage(
			values.GameImageIDFromUUID(image.ID),
			imageType,
			image.CreatedAt,
		),
		GameID: values.NewGameIDFromUUID(image.GameID),
	}, nil
}

func (gameImage *GameImageV2) GetGameImages(ctx context.Context, gameID values.GameID, lockType repository.LockType) ([]*domain.GameImage, error) {
	db, err := gameImage.db.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get db: %w", err)
	}

	db, err = gameImage.db.setLock(db, lockType)
	if err != nil {
		return nil, fmt.Errorf("failed to set lock: %w", err)
	}

	var images []schema.GameImageTable2
	err = db.
		Joins("GameImageType").
		Where("game_id = ?", uuid.UUID(gameID)).
		Order("created_at DESC").
		Find(&images).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game images: %w", err)
	}

	gameImages := make([]*domain.GameImage, 0, len(images))
	for _, image := range images {
		var imageType values.GameImageType
		switch image.GameImageType.Name {
		case schema.GameImageTypeJpeg:
			imageType = values.GameImageTypeJpeg
		case schema.GameImageTypePng:
			imageType = values.GameImageTypePng
		case schema.GameImageTypeGif:
			imageType = values.GameImageTypeGif
		default:
			return nil, fmt.Errorf("invalid image type: %s", image.GameImageType.Name)
		}

		gameImages = append(gameImages, domain.NewGameImage(
			values.GameImageIDFromUUID(image.ID),
			imageType,
			image.CreatedAt,
		))
	}

	return gameImages, nil
}
