package migrate

// 注意:
// 以前のマイグレーションとの互換性を保つために、
// 他のバージョンとは違いテーブル名にV1のようなバージョンをつけない

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
		&gameTable{},
		&gameVersionTable{},
		&gameURLTable{},
		&gameFileTable{},
		&gameFileTypeTable{},
		&gameImageTable{},
		&gameImageTypeTable{},
		&gameVideoTable{},
		&gameVideoTypeTable{},
		&gameManagementRoleTable{},
		&gameManagementRoleTypeTable{},
		&launcherVersionTable{},
		&launcherUserTable{},
		&launcherSessionTable{},
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

type gameTable struct {
	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
	Description         string                    `gorm:"type:text;not null"`
	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
	GameVersions        []gameVersionTable        `gorm:"foreignkey:GameID"`
	GameManagementRoles []gameManagementRoleTable `gorm:"foreignKey:GameID"`
	GameImages          []gameImageTable          `gorm:"foreignKey:GameID"`
	GameVideos          []gameVideoTable          `gorm:"foreignKey:GameID"`
}

func (*gameTable) TableName() string {
	return "games"
}

type gameVersionTable struct {
	ID          uuid.UUID       `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID       `gorm:"type:varchar(36);not null"`
	Name        string          `gorm:"type:varchar(32);size:32;not null"`
	Description string          `gorm:"type:text;not null"`
	CreatedAt   time.Time       `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFiles   []gameFileTable `gorm:"foreignKey:GameVersionID"`
	GameURL     gameURLTable    `gorm:"foreignKey:GameVersionID"`
}

func (*gameVersionTable) TableName() string {
	return "game_versions"
}

type gameURLTable struct {
	ID            uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameVersionID uuid.UUID `gorm:"type:varchar(36);not null;unique"`
	URL           string    `gorm:"type:text;not null"`
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (*gameURLTable) TableName() string {
	return "game_urls"
}

type gameFileTable struct {
	ID            uuid.UUID         `gorm:"type:varchar(36);not null;primaryKey"`
	GameVersionID uuid.UUID         `gorm:"type:varchar(36);not null;index:idx_game_file_unique,unique"`
	FileTypeID    int               `gorm:"type:tinyint;not null;index:idx_game_file_unique,unique"`
	Hash          string            `gorm:"type:char(32);size:32;not null"`
	EntryPoint    string            `gorm:"type:text;not null"`
	CreatedAt     time.Time         `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFileType  gameFileTypeTable `gorm:"foreignKey:FileTypeID"`
}

func (*gameFileTable) TableName() string {
	return "game_files"
}

const (
	gameFileTypeJarV1     = "jar"
	gameFileTypeWindowsV1 = "windows"
	gameFileTypeMacV1     = "mac"
)

type gameFileTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameFileTypeTable) TableName() string {
	return "game_file_types"
}

func setupGameFileTypeTableV1(db *gorm.DB) error {
	fileTypes := []gameFileTypeTable{
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

type gameImageTable struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	ImageTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameImageType gameImageTypeTable `gorm:"foreignKey:ImageTypeID"`
}

func (*gameImageTable) TableName() string {
	return "game_images"
}

const (
	gameImageTypeJpegV1 = "jpeg"
	gameImageTypePngV1  = "png"
	gameImageTypeGifV1  = "gif"
)

type gameImageTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameImageTypeTable) TableName() string {
	return "game_image_types"
}

func setupGameImageTypeTableV1(db *gorm.DB) error {
	imageTypes := []gameImageTypeTable{
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

type gameVideoTable struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	VideoTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVideoType gameVideoTypeTable `gorm:"foreignKey:VideoTypeID"`
}

func (*gameVideoTable) TableName() string {
	return "game_videos"
}

const (
	gameVideoTypeMp4V1 = "mp4"
)

type gameVideoTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameVideoTypeTable) TableName() string {
	return "game_video_types"
}

func setupGameVideoTypeTableV1(db *gorm.DB) error {
	videoTypes := []gameVideoTypeTable{
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

type gameManagementRoleTable struct {
	GameID        uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	UserID        uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	RoleTypeID    int                         `gorm:"type:tinyint;not null"`
	RoleTypeTable gameManagementRoleTypeTable `gorm:"foreignKey:RoleTypeID"`
}

func (*gameManagementRoleTable) TableName() string {
	return "game_management_roles"
}

const (
	gameManagementRoleTypeAdministratorV1 = "administrator"
	gameManagementRoleTypeCollaboratorV1  = "collaborator"
)

type gameManagementRoleTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*gameManagementRoleTypeTable) TableName() string {
	return "game_management_role_types"
}

func setupGameManagementRoleTypeTableV1(db *gorm.DB) error {
	roleTypes := []gameManagementRoleTypeTable{
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

type launcherVersionTable struct {
	ID               uuid.UUID           `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string              `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString      `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time           `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt      `gorm:"type:DATETIME NULL;default:NULL"`
	LauncherUsers    []launcherUserTable `gorm:"foreignKey:LauncherVersionID"`
	// gormigrateを使用していなかったv1との互換性のため、
	// joinForeignKey、joinReferencesを指定している
	Games []gameTable `gorm:"many2many:launcher_version_game_relations"`
}

func (*launcherVersionTable) TableName() string {
	return "launcher_versions"
}

type launcherUserTable struct {
	ID                uuid.UUID              `gorm:"type:varchar(36);not null;primaryKey"`
	LauncherVersionID uuid.UUID              `gorm:"type:varchar(36);not null"`
	ProductKey        string                 `gorm:"type:varchar(29);not null;unique"`
	CreatedAt         time.Time              `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt         gorm.DeletedAt         `gorm:"type:DATETIME NULL;default:NULL"`
	LauncherSessions  []launcherSessionTable `gorm:"foreignKey:LauncherUserID"`
}

func (*launcherUserTable) TableName() string {
	return "launcher_users"
}

type launcherSessionTable struct {
	ID             uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
	LauncherUserID uuid.UUID      `gorm:"type:varchar(36);not null"`
	AccessToken    string         `gorm:"type:varchar(64);not null;unique"`
	ExpiresAt      time.Time      `gorm:"type:datetime;not null"`
	CreatedAt      time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
}

func (lst *launcherSessionTable) TableName() string {
	return "launcher_sessions"
}
