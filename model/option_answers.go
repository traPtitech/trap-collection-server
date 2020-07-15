package model

// OptionAnswer 選択肢式の回答の構造体
type OptionAnswer struct {
	ID             uint   `gorm:"type:int(11) unsigned auto_increment;not null;primary_key;"`
	ResponseID     string `gorm:"type:varchar(36);not null;"`
	Response       Response
	QuestionID     uint `gorm:"type:int(11) unsigned;not null;"`
	Question       Question
	OptionID       uint           `gorm:"type:int(11) unsigned;not null;"`
	QuestionOption QuestionOption `gorm:"foreign_key:OptionID;"`
}
