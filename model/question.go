package model

import "time"

// Question 質問の構造体
type Question struct {
	ID uint `gorm:"type:int(11) unsigned;auto_increament;primary_key;"`
	LauncherVersionID uint `gorm:"type:int(11) unsigned;not null;"`
	QuestionNum uint `gorm:"type:int(11) unsinged;not null;"`
	Type uint8 `gorm:"type:tinyint;not null;"`
	Contents string `gorm:"type:text;not null;"`
	Required bool `gorm:"type:boolean;not null;default:true;"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
	DeletedAt time.Time `gorm:"type:datetime;default:null;"`
}

// QuestionOption 選択肢の構造体
type QuestionOption struct {
	ID uint `gorm:"type:int(11) unsigned;not null;primary_key;auto_increament;"`
	QuestionID uint `gorm:"type:int (11) unsigned;not null"`
	Question Question
	Label string `gorm:"type:text;not null;"`
}