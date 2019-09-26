package model

import (
	"time"
)

//GameTime 各ゲームのプレイ時間の構造体
type GameTime struct {
	GameID    string    `json:"game_id,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

//Time ゲームのプレイ時間
type Time struct {
	VersionID string     `json:"version_id,omitempty"`
	List      []GameTime `json:"list,omitempty"`
}
