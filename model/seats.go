package model

////go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"time"

	"github.com/traPtitech/trap-collection-server/openapi"
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
	GetSeatDetails(seatVersionID string) ([]*openapi.SeatDetail, error)
}
