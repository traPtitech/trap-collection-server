package model

import (
	"database/sql"
	"time"
)

// AccessToken access tokenの構造体
type AccessToken struct {
	ID          string       `gorm:"type:varchar(36);PRIMARY_KEY;"`
	KeyID       string       `gorm:"type:varchar(36);NOT NULL;"`
	AccessToken string       `gorm:"type:varchar(36);NOT NULL;"`
	ExpiresIn   time.Time    `gorm:"type:int(11);NOT NULL;"`
	CreatedAt   time.Time    `gorm:"type:datetime;NOT NULL;"`
	DeletedAt   sql.NullTime `gorm:"type:datetime;DEFAULT:NULL;"`
}
