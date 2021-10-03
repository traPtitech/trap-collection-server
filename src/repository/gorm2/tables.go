package gorm2

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	tables = []interface{}{
		&LauncherUserTable{},
		&LauncherSessionTable{},
	}
)

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
