package model

////go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// Seat プレイヤーの履歴の構造体
type Seat struct {
	ID            string    `gorm:"type:varchar(36);NOT NULL;PRIMARY_KEY;"`
	SeatVersionID string    `gorm:"type:varchar(36);not null;"`
	Row           uint      `gorm:"type:int(11) unsigned;not null;"`
	Column        uint      `gorm:"type:int(11) unsigned;not null;"`
	StartedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
	EndedAt       time.Time `gorm:"type:datetime;default:null;"`
}

type SeatMeta interface {
	GetSeats(seatVersionID string) ([]*Seat, error)
}

func (*DB) GetSeats(seatVersionID string) ([]*Seat, error) {
	var seats []*Seat
	err := db.
		Where("seat_version_id = ? AND ended_at IS NULL", seatVersionID).
		Find(&seats).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("failed to get seats: %w", err)
	}

	return seats, nil
}
