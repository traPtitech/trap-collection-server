package model

////go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"time"
)

// Seat プレイヤーの履歴の構造体
type Seat struct {
	ID            uint      `gorm:"type:int(11) unsigned auto_increment;NOT NULL;PRIMARY_KEY;"`
	SeatVersionID uint      `gorm:"type:int(11) unsigned;not null;"`
	Row           uint      `gorm:"type:int(11) unsigned;not null;"`
	Column        uint      `gorm:"type:int(11) unsigned;not null;"`
	StartedAt     time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
	EndedAt       time.Time `gorm:"type:datetime;default:null;"`
}
