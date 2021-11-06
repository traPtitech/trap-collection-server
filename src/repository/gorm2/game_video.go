package gorm2

import (
	"context"
	"fmt"

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
			Name: gameVideoTypeMp4,
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
