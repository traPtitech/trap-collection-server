package model

import (
	"time"
)

//VersionForSale 販売用のバージョンの構造体
type VersionForSale struct {
	ID          string    `json:"id,omitempty" db:"id"`
	Name        string    `json:"name,omitempty" db:"name"`
	StartPeriod time.Time `json:"start_period,omitempty" db:"start_period"`
	EndPeriod   time.Time `json:"end_period,omitempty" db:"end_period"`
	StartTime   time.Time `json:"start_time,omitempty" db:"start_time"`
	CreatedAt   time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

//VersionNotForSale 工大祭用のバージョンの構造体
type VersionNotForSale struct {
	ID              string    `json:"id,omitempty" db:"id"`
	Name            string    `json:"name,omitempty" db:"name"`
	QuestionnaireID int       `json:"questionnaire_id" db:"questionnaire_id"`
	StartPeriod     time.Time `json:"start_period,omitempty" db:"start_period"`
	EndPeriod       time.Time `json:"end_period,omitempty" db:"end_period"`
	StartTime       time.Time `json:"start_time,omitempty" db:"start_time"`
	CreatedAt       time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

//Special 特例の構造体
type Special struct {
	GameName string `json:"game_name,omitempty"`
	InOut    string `json:"in_out,omitempty"`
}
