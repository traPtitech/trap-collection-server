package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository/gorm2/migrate"
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
		imageTypeName = migrate.GameImageTypeJpeg
	case values.GameImageTypePng:
		imageTypeName = migrate.GameImageTypePng
	case values.GameImageTypeGif:
		imageTypeName = migrate.GameImageTypeGif
	default:
		return fmt.Errorf("invalid image type: %d", image.GetType())
	}

	var imageType migrate.GameImageTypeTable
	err = db.
		Where("name = ?", imageTypeName).
		Select("id").
		Take(&imageType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	imageTypeID := imageType.ID

	err = db.
		Create(&migrate.GameImageTable2{
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
