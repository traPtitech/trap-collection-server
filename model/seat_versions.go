package model

import (
	"database/sql"
	"time"
)

// SeatVersion seat_versionsの構造体
type SeatVersion struct {
	ID                int          `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	LauncherVersionID int          `gorm:"type:int(11);NOT NULL;"`
	CreatedAt         time.Time    `gorm:"type:datetime;NOT NULL;"`
	DeletedAt         sql.NullTime `gorm:"type:datetime;DEFAULT:NULL;"`
}
