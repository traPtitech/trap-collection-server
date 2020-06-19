package model

// TextAnswer テキスト形式の回答の構造体
type TextAnswer struct {
	ID         uint   `gorm:"type:int(11) unsigned auto_increment;not null;primary_key;"`
	ResponseID string `gorm:"type:varchar(36);not null;"`
	Response   Response
	QuestionID uint `gorm:"type:int(11) unsigned;not null;"`
	Question   Question
	Content    string `gorm:"type:text;not null;"`
}
