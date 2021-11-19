package gorm2

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

const (
	gameFileTypeJar     = "jar"
	gameFileTypeWindows = "windows"
	gameFileTypeMac     = "mac"
)

type GameFile struct {
	db *DB
}

var fileTypeSetupGroup = &singleflight.Group{}

func NewGameFile(db *DB) (*GameFile, error) {
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
	_, err, _ = fileTypeSetupGroup.Do("setupFileTypeTable", func() (interface{}, error) {
		return nil, setupFileTypeTable(gormDB)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to setup file type table: %w", err)
	}

	return &GameFile{
		db: db,
	}, nil
}

func setupFileTypeTable(db *gorm.DB) error {
	fileTypes := []GameFileTypeTable{
		{
			Name:   gameFileTypeJar,
			Active: true,
		},
		{
			Name:   gameFileTypeWindows,
			Active: true,
		},
		{
			Name:   gameFileTypeMac,
			Active: true,
		},
	}

	for _, fileType := range fileTypes {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", fileType.Name).
			FirstOrCreate(&fileType).Error
		if err != nil {
			return fmt.Errorf("failed to create role type: %w", err)
		}
	}

	return nil
}

func (gf *GameFile) SaveGameFile(ctx context.Context, gameVersionID values.GameVersionID, gameFile *domain.GameFile) error {
	gormDB, err := gf.db.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}

	var fileTypeName string
	switch gameFile.GetFileType() {
	case values.GameFileTypeJar:
		fileTypeName = gameFileTypeJar
	case values.GameFileTypeWindows:
		fileTypeName = gameFileTypeWindows
	case values.GameFileTypeMac:
		fileTypeName = gameFileTypeMac
	default:
		return fmt.Errorf("invalid file type: %d", gameFile.GetFileType())
	}

	var fileType GameFileTypeTable
	err = gormDB.
		Where("name = ?", fileTypeName).
		Where("active").
		Select("id").
		Take(&fileType).Error
	if err != nil {
		return fmt.Errorf("failed to get role type: %w", err)
	}
	fileTypeID := fileType.ID

	err = gormDB.Create(&GameFileTable{
		ID:            uuid.UUID(gameFile.GetID()),
		GameVersionID: uuid.UUID(gameVersionID),
		FileTypeID:    fileTypeID,
		Hash:          []byte(gameFile.GetHash()),
		EntryPoint:    string(gameFile.GetEntryPoint()),
	}).Error
	if err != nil {
		return fmt.Errorf("failed to create game image: %w", err)
	}

	return nil
}
