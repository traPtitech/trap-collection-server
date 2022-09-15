package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

var (
	tables = []any{
		&GameTable{},
		&GameVersionTable{},
		&GameURLTable{},
		&GameFileTable{},
		&GameFileTypeTable{},
		&GameImageTable{},
		&GameImageTypeTable{},
		&GameVideoTable{},
		&GameVideoTypeTable{},
		&GameManagementRoleTable{},
		&GameManagementRoleTypeTable{},
		&LauncherVersionTable{},
		&LauncherUserTable{},
		&LauncherSessionTable{},
	}
)

// アプリケーションのv1
type (
	GameTable                   = gameTableV1
	GameVersionTable            = gameVersionTableV1
	GameURLTable                = gameURLTableV1
	GameFileTable               = gameFileTableV1
	GameFileTypeTable           = gameFileTypeTableV1
	GameImageTable              = gameImageTableV1
	GameImageTypeTable          = gameImageTypeTableV1
	GameVideoTable              = gameVideoTableV1
	GameVideoTypeTable          = gameVideoTypeTableV1
	GameManagementRoleTable     = gameManagementRoleTableV1
	GameManagementRoleTypeTable = gameManagementRoleTypeTableV1
	LauncherVersionTable        = launcherVersionTableV1
	LauncherUserTable           = launcherUserTableV1
	LauncherSessionTable        = launcherSessionTableV1
)

const (
	GameFileTypeJar     = gameFileTypeJarV1
	GameFileTypeWindows = gameFileTypeWindowsV1
	GameFileTypeMac     = gameFileTypeMacV1
)

func setupGameFileTypeTable(db *gorm.DB) error {
	fileTypes := []GameFileTypeTable{
		{
			Name:   GameFileTypeJar,
			Active: true,
		},
		{
			Name:   GameFileTypeWindows,
			Active: true,
		},
		{
			Name:   GameFileTypeMac,
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

const (
	GameImageTypeJpeg = gameImageTypeJpegV1
	GameImageTypePng  = gameImageTypePngV1
	GameImageTypeGif  = gameImageTypeGifV1
)

func setupGameImageTypeTable(db *gorm.DB) error {
	imageTypes := []GameImageTypeTable{
		{
			Name:   GameImageTypeJpeg,
			Active: true,
		},
		{
			Name:   GameImageTypePng,
			Active: true,
		},
		{
			Name:   GameImageTypeGif,
			Active: true,
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

const (
	GameVideoTypeMp4 = gameVideoTypeMp4V1
)

func setupGameVideoTypeTable(db *gorm.DB) error {
	videoTypes := []GameVideoTypeTable{
		{
			Name:   GameVideoTypeMp4,
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

const (
	GameManagementRoleTypeAdministrator = gameManagementRoleTypeAdministratorV1
	GameManagementRoleTypeCollaborator  = gameManagementRoleTypeCollaboratorV1
)

func setupGameManagementRoleTypeTable(db *gorm.DB) error {
	roleTypes := []GameManagementRoleTypeTable{
		{
			Name:   GameManagementRoleTypeAdministrator,
			Active: true,
		},
		{
			Name:   GameManagementRoleTypeCollaborator,
			Active: true,
		},
	}

	for _, roleType := range roleTypes {
		err := db.
			Session(&gorm.Session{}).
			Where("name = ?", roleType.Name).
			FirstOrCreate(&roleType).Error
		if err != nil {
			return fmt.Errorf("failed to create role type: %w", err)
		}
	}

	return nil
}
