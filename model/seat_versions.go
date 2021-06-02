package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// SeatVersion seat_versionsの構造体
type SeatVersion struct {
	ID        string       `gorm:"type:varchar(36);PRIMARY_KEY;"`
	Height    uint         `gorm:"type:int(11) unsigned;not null;"`
	Width     uint         `gorm:"type:int(11) unsigned;not null;"`
	CreatedAt time.Time    `gorm:"type:datetime;NOT NULL;"`
	DeletedAt sql.NullTime `gorm:"type:datetime;DEFAULT:NULL;"`
}

type SeatVersionMeta interface {
	InsertSeatVersion(height uint, width uint) (*openapi.SeatVersion, error)
}

func (*DB) InsertSeatVersion(height uint, width uint) (*openapi.SeatVersion, error) {
	id := uuid.New().String()

	seatVersion := SeatVersion{
		ID: id,
		Height: height,
		Width: width,
	}

	err := db.Create(&seatVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to INSERT seat version record")
	}

	return &openapi.SeatVersion{
		Id: seatVersion.ID,
		Width: int32(seatVersion.Width),
		Hight: int32(seatVersion.Height),
		CreatedAt: seatVersion.CreatedAt,
	}, nil
}
