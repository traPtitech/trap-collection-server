package model

import (
	"time"
)

// Question 質問の構造体
type Question struct {
	ID uint `gorm:"type:int(11) unsigned;auto_increament;primary_key;"`
	LauncherVersionID uint `gorm:"type:int(11) unsigned;not null;"`
	QuestionNum uint `gorm:"type:int(11) unsigned;not null;"`
	Type uint8 `gorm:"type:tinyint unsigned;not null;"`
	Content string `gorm:"type:text;not null;"`
	Required bool `gorm:"type:boolean;not null;default:true;"`
	QuestionOptions []QuestionOption `gorm:"foreign_key:QuestionID;"`
	CreatedAt time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
	DeletedAt time.Time `gorm:"type:datetime;default:null;"`
}

var typeMap = map[uint8]string{
	0:"radio",
	1:"checkbox",
	2:"text",
}

// QuestionOption 選択肢の構造体
type QuestionOption struct {
	ID uint `gorm:"type:int(11) unsigned;not null;primary_key;auto_increament;"`
	QuestionID uint `gorm:"type:int (11) unsigned;not null"`
	Question Question
	Label string `gorm:"type:text;not null;"`
}