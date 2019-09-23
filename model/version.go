package model

import (
	"time"
)

//VersionForSale 販売用のバージョンの構造体
type VersionForSale struct {
	Name string `json:"name" db:"name"`
	StartPeriod time.Time `json:"start_period"`
	EndPeriod time.Time `json:"end_period" db:"end_period"`
	StartTime time.Time `json:"start_time" db:"start_time"`
}

//VersionNotForSale 工大祭用のバージョンの構造体
type VersionNotForSale struct {
	Name string `json:"name" db:"name"`
	QuestionnaireID int `json:"questionnaire_id" db:"questionnaire_id"`
	StartPeriod time.Time `json:"start_period"`
	EndPeriod time.Time `json:"end_period" db:"end_period"`
	StartTime time.Time `json:"start_time" db:"start_time"`
}
