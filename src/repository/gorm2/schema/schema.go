package schema

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// type GameTable struct {
// 	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
// 	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
// 	Description         string                    `gorm:"type:text;not null"`
// 	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
// 	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
// 	GameVersions        []GameVersionTable        `gorm:"foreignkey:GameID"`
// 	GameManagementRoles []GameManagementRoleTable `gorm:"foreignKey:GameID"`
// 	GameImages          []GameImageTable          `gorm:"foreignKey:GameID"`
// 	GameVideos          []GameVideoTable          `gorm:"foreignKey:GameID"`
// }

// func (*GameTable) TableName() string {
// 	return "games"
// }

type GameVersionTable struct {
	ID          uuid.UUID       `gorm:"type:varchar(36);not null;primaryKey"`
	GameID      uuid.UUID       `gorm:"type:varchar(36);not null"`
	Name        string          `gorm:"type:varchar(32);size:32;not null"`
	Description string          `gorm:"type:text;not null"`
	CreatedAt   time.Time       `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFiles   []GameFileTable `gorm:"foreignKey:GameVersionID"`
	GameURL     GameURLTable    `gorm:"foreignKey:GameVersionID"`
}

func (*GameVersionTable) TableName() string {
	return "game_versions"
}

type GameURLTable struct {
	ID            uuid.UUID `gorm:"type:varchar(36);not null;primaryKey"`
	GameVersionID uuid.UUID `gorm:"type:varchar(36);not null;unique"`
	URL           string    `gorm:"type:text;not null"`
	CreatedAt     time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (*GameURLTable) TableName() string {
	return "game_urls"
}

type GameFileTable struct {
	ID            uuid.UUID         `gorm:"type:varchar(36);not null;primaryKey"`
	GameVersionID uuid.UUID         `gorm:"type:varchar(36);not null;index:idx_game_file_unique,unique"`
	FileTypeID    int               `gorm:"type:tinyint;not null;index:idx_game_file_unique,unique"`
	Hash          string            `gorm:"type:char(32);size:32;not null"`
	EntryPoint    string            `gorm:"type:text;not null"`
	CreatedAt     time.Time         `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameFileType  GameFileTypeTable `gorm:"foreignKey:FileTypeID"`
}

func (*GameFileTable) TableName() string {
	return "game_files"
}

type GameFileTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*GameFileTypeTable) TableName() string {
	return "game_file_types"
}

type GameImageTable struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	ImageTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameImageType GameImageTypeTable `gorm:"foreignKey:ImageTypeID"`
}

func (*GameImageTable) TableName() string {
	return "game_images"
}

type GameImageTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*GameImageTypeTable) TableName() string {
	return "game_image_types"
}

type GameVideoTable struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	VideoTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameVideoType GameVideoTypeTable `gorm:"foreignKey:VideoTypeID"`
}

func (*GameVideoTable) TableName() string {
	return "game_videos"
}

type GameVideoTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*GameVideoTypeTable) TableName() string {
	return "game_video_types"
}

type GameManagementRoleTable struct {
	GameID        uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	UserID        uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	RoleTypeID    int                         `gorm:"type:tinyint;not null"`
	RoleTypeTable GameManagementRoleTypeTable `gorm:"foreignKey:RoleTypeID"`
}

func (*GameManagementRoleTable) TableName() string {
	return "game_management_roles"
}

type GameManagementRoleTypeTable struct {
	ID     int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name   string `gorm:"type:varchar(32);size:32;not null;unique"`
	Active bool   `gorm:"type:boolean;default:true"`
}

func (*GameManagementRoleTypeTable) TableName() string {
	return "game_management_role_types"
}

type LauncherVersionTable struct {
	ID               uuid.UUID           `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string              `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString      `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time           `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt      `gorm:"type:DATETIME NULL;default:NULL"`
	LauncherUsers    []LauncherUserTable `gorm:"foreignKey:LauncherVersionID"`
	// gormigrateを使用していなかったv1との互換性のため、
	// joinForeignKey、joinReferencesを指定している
	Games []GameTable2V15 `gorm:"many2many:launcher_version_game_relations"`
}

func (*LauncherVersionTable) TableName() string {
	return "launcher_versions"
}

type LauncherUserTable struct {
	ID                uuid.UUID              `gorm:"type:varchar(36);not null;primaryKey"`
	LauncherVersionID uuid.UUID              `gorm:"type:varchar(36);not null"`
	ProductKey        string                 `gorm:"type:varchar(29);not null;unique"`
	CreatedAt         time.Time              `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt         gorm.DeletedAt         `gorm:"type:DATETIME NULL;default:NULL"`
	LauncherSessions  []LauncherSessionTable `gorm:"foreignKey:LauncherUserID"`
}

func (*LauncherUserTable) TableName() string {
	return "launcher_users"
}

type LauncherSessionTable struct {
	ID             uuid.UUID      `gorm:"type:varchar(36);not null;primaryKey"`
	LauncherUserID uuid.UUID      `gorm:"type:varchar(36);not null"`
	AccessToken    string         `gorm:"type:varchar(64);not null;unique"`
	ExpiresAt      time.Time      `gorm:"type:datetime;not null"`
	CreatedAt      time.Time      `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt `gorm:"type:DATETIME NULL;default:NULL"`
}

func (lst *LauncherSessionTable) TableName() string {
	return "launcher_sessions"
}
