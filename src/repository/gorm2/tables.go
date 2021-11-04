package gorm2

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	tables = []interface{}{
		&GameTable{},
		&GameImageTable{},
		&GameImageTypeTable{},
		&GameManagementRoleTable{},
		&GameManagementRoleTypeTable{},
		&LauncherVersionTable{},
		&LauncherUserTable{},
		&LauncherSessionTable{},
	}
)

type GameTable struct {
	ID                  uuid.UUID                 `gorm:"type:varchar(36);not null;primaryKey"`
	Name                string                    `gorm:"type:varchar(256);size:256;not null"`
	Description         string                    `gorm:"type:text;not null"`
	CreatedAt           time.Time                 `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt            `gorm:"type:DATETIME NULL;default:NULL"`
	GameManagementRoles []GameManagementRoleTable `gorm:"foreignKey:GameID"`
	GameImages          []GameImageTable          `gorm:"foreignKey:GameID"`
}

func (gt *GameTable) TableName() string {
	return "games"
}

type GameImageTable struct {
	ID            uuid.UUID          `gorm:"type:varchar(36);not null;primaryKey"`
	GameID        uuid.UUID          `gorm:"type:varchar(36);not null"`
	ImageTypeID   int                `gorm:"type:tinyint;not null"`
	CreatedAt     time.Time          `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	GameImageType GameImageTypeTable `gorm:"foreignKey:ImageTypeID"`
}

func (gt *GameImageTable) TableName() string {
	return "game_images"
}

type GameImageTypeTable struct {
	ID   int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name string `gorm:"type:varchar(32);size:32;not null;unique"`
}

func (gtt *GameImageTypeTable) TableName() string {
	return "game_image_types"
}

type GameManagementRoleTable struct {
	GameID        uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	UserID        uuid.UUID                   `gorm:"type:varchar(36);not null;primaryKey"`
	RoleTypeID    int                         `gorm:"type:tinyint;not null"`
	RoleTypeTable GameManagementRoleTypeTable `gorm:"foreignKey:RoleTypeID"`
}

func (gmrt *GameManagementRoleTable) TableName() string {
	return "game_management_roles"
}

type GameManagementRoleTypeTable struct {
	ID   int    `gorm:"type:TINYINT AUTO_INCREMENT;not null;primaryKey"`
	Name string `gorm:"type:varchar(32);size:32;not null;unique"`
}

func (gmrt *GameManagementRoleTypeTable) TableName() string {
	return "game_management_role_types"
}

type LauncherVersionTable struct {
	ID               uuid.UUID           `gorm:"type:varchar(36);not null;primaryKey"`
	Name             string              `gorm:"type:varchar(32);not null;unique"`
	QuestionnaireURL sql.NullString      `gorm:"type:text;default:NULL"`
	CreatedAt        time.Time           `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt      `gorm:"type:DATETIME NULL;default:NULL"`
	LauncherUsers    []LauncherUserTable `gorm:"foreignKey:LauncherVersionID"`
}

func (lvt *LauncherVersionTable) TableName() string {
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

func (lut *LauncherUserTable) TableName() string {
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
