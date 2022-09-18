package migrate

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v1
// アプリケーションのv1時のマイグレーション
func v1() *gormigrate.Migration {
	tables := []any{
		&gameTableV1{},
		&gameVersionTableV1{},
		&gameURLTableV1{},
		&gameFileTableV1{},
		&gameFileTypeTableV1{},
		&gameImageTableV1{},
		&gameImageTypeTableV1{},
		&gameVideoTableV1{},
		&gameVideoTypeTableV1{},
		&gameManagementRoleTableV1{},
		&gameManagementRoleTypeTableV1{},
		&launcherVersionTableV1{},
		&launcherUserTableV1{},
		&launcherSessionTableV1{},
	}

	return &gormigrate.Migration{
		ID: "1",
		Migrate: func(tx *gorm.DB) error {
			err := tx.AutoMigrate(tables...)
			if err != nil {
				return fmt.Errorf("failed to migrate: %w", err)
			}

			err = setupGameFileTypeTableV1(tx)
			if err != nil {
				return fmt.Errorf("failed to setup game file type table: %w", err)
			}

			err = setupGameImageTypeTableV1(tx)
			if err != nil {
				return fmt.Errorf("failed to setup game image type table: %w", err)
			}

			err = setupGameVideoTypeTableV1(tx)
			if err != nil {
				return fmt.Errorf("failed to setup game video type table: %w", err)
			}

			err = setupGameManagementRoleTypeTableV1(tx)
			if err != nil {
				return fmt.Errorf("failed to setup game management role type table: %w", err)
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(tables...)
		},
	}
}

type gameTableV1 struct {
	ID                  uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                      `gorm:"type:varchar(256);size:256;not null"`
	Description         string                      `gorm:"type:text;not null"`
	CreatedAt           time.Time                   `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt              `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTableV1        `gorm:"foreignkey:GameID"`
	GameManagementRoles []gameManagementRoleTableV1 `gorm:"foreignKey:GameID"`
	GameImages          []gameImageTableV1          `gorm:"foreignKey:GameID"`
	GameVideos          []gameVideoTableV1          `gorm:"foreignKey:GameID"`
}

func (*gameTableV1) TableName() string {
	return "games"
}

type gameVersionTableV1 struct {
	ID          uuid.UUID         `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID         `gorm:"type:varchar(36);not null"`
	Name        string            `gorm:"type:varchar(32);size:32;not null"`
	Description string            `gorm:"type:text;not null"`
	CreatedAt   time.Time         `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFiles   []gameFileTableV1 `gorm:"foreignKey:GameVersionID"`
	GameURL     gameURLTableV1    `gorm:"foreignKey:GameVersionID"`
}

func (*gameVersionTableV1) TableName() string {
	return "game_versions"
}

type gameURLTableV1 struct {
	ID            uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameVersionID uuid.UUID `gorm:"type:varchar(36);not null;unique"`
	URL           string    `gorm:"type:text;not null"`
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (*gameURLTableV1) TableName() string {
	return "game_urls"
}

type gameFileTableV1 struct {
	ID            uuid.UUID           `gorm:"type:varchar(36);not null;primaryKey"`
	GameVersionID uuid.UUID           `gorm:"type:varchar(36);not null;index:idx_game_file_unique,unique"`
	FileTypeID    int                 `gorm:"type:tinyint;not null;index:idx_game_file_unique,unique"`
	Hash          string              `gorm:"type:char(32);size:32;not null"`
	EntryPoint    string              `gorm:"type:text;not null"`
	CreatedAt     time.Time           `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFileType  gameFileTypeTableV1 `gorm:"foreignKey:FileTypeID"`
}

func (*gameFileTableV1) TableName() string {
	return "game_files"
}

const (
	gameFileTypeJarV1     = "jar"
	gameFileTypeWindowsV1 = "windows"
	gameFileTypeMacV1     = "mac"
)

type gameFileTypeTableV1 struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameFileTypeTableV1) TableName() string {
	return "game_file_types"
}

func setupGameFileTypeTableV1(db *gorm.DB) error {
	fileTypes := []gameFileTypeTableV1{
		{
			Name:   gameFileTypeJarV1,
			Active: true,
		},
		{
			Name:   gameFileTypeWindowsV1,
			Active: true,
		},
		{
			Name:   gameFileTypeMacV1,
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

type gameImageTableV1 struct {
	ID            uuid.UUID            `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID            `gorm:"type:varchar(36);not null"`
	ImageTypeID   int                  `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time            `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameImageType gameImageTypeTableV1 `gorm:"foreignKey:ImageTypeID"`
}

func (*gameImageTableV1) TableName() string {
	return "game_images"
}

const (
	gameImageTypeJpegV1 = "jpeg"
	gameImageTypePngV1  = "png"
	gameImageTypeGifV1  = "gif"
)

type gameImageTypeTableV1 struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameImageTypeTableV1) TableName() string {
	return "game_image_types"
}

func setupGameImageTypeTableV1(db *gorm.DB) error {
	imageTypes := []gameImageTypeTableV1{
		{
			Name:   gameImageTypeJpegV1,
			Active: true,
		},
		{
			Name:   gameImageTypePngV1,
			Active: true,
		},
		{
			Name:   gameImageTypeGifV1,
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

type gameVideoTableV1 struct {
	ID            uuid.UUID            `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID            `gorm:"type:varchar(36);not null"`
	VideoTypeID   int                  `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time            `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVideoType gameVideoTypeTableV1 `gorm:"foreignKey:VideoTypeID"`
}

func (*gameVideoTableV1) TableName() string {
	return "game_videos"
}

const (
	gameVideoTypeMp4V1 = "mp4"
)

type gameVideoTypeTableV1 struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameVideoTypeTableV1) TableName() string {
	return "game_video_types"
}

func setupGameVideoTypeTableV1(db *gorm.DB) error {
	videoTypes := []gameVideoTypeTableV1{
		{
			Name:   gameVideoTypeMp4V1,
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

type gameManagementRoleTableV1 struct {
	GameID        uuid.UUID                     `gorm:"type:varchar(36);not null;primaryKey"`
	UserID        uuid.UUID                     `gorm:"type:varchar(36);not null;primaryKey"`
	RoleTypeID    int                           `gorm:"type:tinyint;not null"`
	RoleTypeTable gameManagementRoleTypeTableV1 `gorm:"foreignKey:RoleTypeID"`
}

func (*gameManagementRoleTableV1) TableName() string {
	return "game_management_roles"
}

const (
	gameManagementRoleTypeAdministratorV1 = "administrator"
	gameManagementRoleTypeCollaboratorV1  = "collaborator"
)

type gameManagementRoleTypeTableV1 struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameManagementRoleTypeTableV1) TableName() string {
	return "game_management_role_types"
}

func setupGameManagementRoleTypeTableV1(db *gorm.DB) error {
	roleTypes := []gameManagementRoleTypeTableV1{
		{
			Name:   gameManagementRoleTypeAdministratorV1,
			Active: true,
		},
		{
			Name:   gameManagementRoleTypeCollaboratorV1,
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

type launcherVersionTableV1 struct {
	ID               uuid.UUID             `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string                `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString        `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time             `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt        `gorm:"type:DATETIME NULL;default:NULL"`
	LauncherUsers    []launcherUserTableV1 `gorm:"foreignKey:LauncherVersionID"`
	// gormigrateを使用していなかったv1との互換性のため、
	// joinForeignKey、joinReferencesを指定している
	Games []gameTableV1 `gorm:"many2many:launcher_version_game_relations;joinForeignKey:LauncherVersionTableID;joinReferences:GameTableID"`
}

func (*launcherVersionTableV1) TableName() string {
	return "launcher_versions"
}

type launcherUserTableV1 struct {
	ID                uuid.UUID                `gorm:"type:varchar(36);not null;primaryKey"`
	LauncherVersionID uuid.UUID                `gorm:"type:varchar(36);not null"`
	ProductKey        string                   `gorm:"type:varchar(29);not null;unique"`
	CreatedAt         time.Time                `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt         gorm.DeletedAt           `gorm:"type:DATETIME NULL;default:NULL"`
	LauncherSessions  []launcherSessionTableV1 `gorm:"foreignKey:LauncherUserID"`
}

func (*launcherUserTableV1) TableName() string {
	return "launcher_users"
}

type launcherSessionTableV1 struct {
	ID             uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
	LauncherUserID uuid.UUID      `gorm:"type:varchar(36);not null"`
	AccessToken    string         `gorm:"type:varchar(64);not null;unique"`
	ExpiresAt      time.Time      `gorm:"type:datetime;not null"`
	CreatedAt      time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
}

func (lst *launcherSessionTableV1) TableName() string {
	return "launcher_sessions"
}
