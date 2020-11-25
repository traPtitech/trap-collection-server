package model

// QuestionOption 選択肢の構造体
type QuestionOption struct {
	ID         uint `gorm:"type:int(11) unsigned auto_increment;not null;primary_key;"`
	QuestionID uint `gorm:"type:int (11) unsigned;not null"`
	Question   Question
	Label      string `gorm:"type:text;not null;"`
}
