package gorm2

import (
	"context"
	"fmt"

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
