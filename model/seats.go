package model

////go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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
	InsertSeat(seatVersionID string, row int, column int) (*Seat, error)
	GetSeats(seatVersionID string) ([]*Seat, error)
}

func (*DB) InsertSeat(seatVersionID string, row int, column int) (*Seat, error) {
	//TODO:同時に2リクエストくると２重に着席できてしまう
	newSeat := Seat{
		ID: uuid.New().String(),
		SeatVersionID: seatVersionID,
		Row: uint(row),
		Column: uint(column),
		StartedAt: time.Now(),
	}

	err := db.
		Where("`row` = ? AND `column` = ? AND ended_at IS NULL", row, column).
		First(&Seat{}).Error
	if err == nil {
		return nil, ErrAlreadyExists
	}
	if !gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("failed to get seat: %w", err)
	}

	err = db.Create(&newSeat).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create seat: %w", err)
	}

	return &newSeat, nil
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
