package model

// GameRating ゲームの評価の構造体
type GameRating struct {
	ID            uint `gorm:"type:int(11) unsigned auto_increment;not null;primary_key;"`
	ResponseID    uint `gorm:"type:varchar(36);not null;"`
	Response      Response
	GameVersionID uint `gorm:"type:int(11) unsigned;not null;"`
	GameVersion   GameVersion
	Star          uint8 `gorm:"type:tinyint unsigned;not null;"`
	PlayTime      uint  `gorm:"type:int(11) unsigned;not null;"`
}
