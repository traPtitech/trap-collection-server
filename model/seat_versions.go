package model

import (
	"database/sql"
	"time"
)

// SeatVersion seat_versionsの構造体
type SeatVersion struct {
	ID        string       `gorm:"type:varchar(36);PRIMARY_KEY;"`
	Height    uint         `gorm:"type:int(11) unsigned;not null;"`
	Width     uint         `gorm:"type:int(11) unsigned;not null;"`
	CreatedAt time.Time    `gorm:"type:datetime;NOT NULL;"`
	DeletedAt sql.NullTime `gorm:"type:datetime;DEFAULT:NULL;"`
}
