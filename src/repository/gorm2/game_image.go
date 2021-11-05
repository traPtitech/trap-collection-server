package gorm2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"golang.org/x/sync/singleflight"
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

var imageTypeSetupGroup = &singleflight.Group{}

func NewGameImage(db *DB) (*GameImage, error) {
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
	_, err, _ = imageTypeSetupGroup.Do("setupImageTypeTable", func() (interface{}, error) {
		return nil, setupImageTypeTable(gormDB)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to setup image type table: %w", err)
	}

	return &GameImage{
		db: db,
	}, nil
}

func setupImageTypeTable(db *gorm.DB) error {
	imageTypes := []GameImageTypeTable{
		{
			Name: gameImageTypeJpeg,
		},
		{
			Name: gameImageTypePng,
		},
		{
			Name: gameImageTypeGif,
		},
	}

	for _, imageType := range imageTypes {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", imageType.Name).
			FirstOrCreate(&imageType).Error
		if err != nil {
			return fmt.Errorf("failed to create role type: %w", err)
		}
	}

	return nil
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

	var imageType GameImageTypeTable
	err = gormDB.
		Where("name = ?", imageTypeName).
		Select("id").
		First(&imageType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	imageTypeID := imageType.ID

	err = gormDB.Create(&GameImageTable{
		ID:          uuid.UUID(image.GetID()),
		GameID:      uuid.UUID(gameID),
		ImageTypeID: imageTypeID,
		CreatedAt:   time.Now(),
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

	var image GameImageTable
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
	), nil
}
