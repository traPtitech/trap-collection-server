package model

import "time"

// Player プレイヤーの履歴の構造体
type Player struct {
	ID        uint      `gorm:"type:int(11) unsigned;NOT NULL;PRIMARY_KEY;AUTO_INCREMENT;"`
	SeatID    uint      `gorm:"type:int(11) unsigned;not null;"`
	StartedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
	EndedAt   time.Time `gorm:"type:datetime;default:null;"`
}
