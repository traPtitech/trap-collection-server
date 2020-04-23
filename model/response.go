package model

import "time"

// Response 回答の構造体
type Response struct {
	ID                string `gorm:"type:varchar(36);not null;primary_key;"`
	PlayerID          uint   `gorm:"type:int(11);not null;"`
	Player            Player
	LauncherVersionID uint `gorm:"type:int(11) unsigned;not null;"`
	LauncherVersion   LauncherVersion
	Remark            string    `gorm:"type:text;"`
	CreatedAt         time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
}

// TextAnswer テキスト形式の回答の構造体
type TextAnswer struct {
	ID         uint   `gorm:"type:int(11) unsigned;not null;primary_key;auto_increment;"`
	ResponseID string `gorm:"type:varchar(36);not null;"`
	Response   Response
	QuestionID uint `gorm:"type:int(11) unsigned;not null;"`
	Question   Question
	Content    string `gorm:"type:text;not null;"`
}

// OptionAnswer 選択肢式の回答の構造体
type OptionAnswer struct {
	ID             uint   `gorm:"type:int(11) unsigned;not null;primary_key;auto_increment;"`
	ResponseID     string `gorm:"type:varchar(36);not null;"`
	Response       Response
	QuestionID     uint `gorm:"type:int(11) unsigned;not null;"`
	Question       Question
	OptionID       uint           `gorm:"type:int(11) unsigned;not null;"`
	QuestionOption QuestionOption `gorm:"foreign_key:OptionID;"`
}

// GameRating ゲームの評価の構造体
type GameRating struct {
	ID            uint `gorm:"type:int(11) unsigned;not null;primary_key;auto_increment;"`
	ResponseID    uint `gorm:"type:varchar(36);not null;"`
	Response      Response
	GameVersionID uint `gorm:"type:int(11) unsigned;not null;"`
	GameVersion   GameVersion
	Star          uint8 `gorm:"type:tinyint unsigned;not null;"`
	PlayTime      uint  `gorm:"type:int(11) unsigned;not null;"`
}
