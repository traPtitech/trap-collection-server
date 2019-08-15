package model

import (
	"time"
)

//Game データベースに格納するgameの構造体
type Game struct {
	GameID    string    `json:"gameId,omitempty" db:"game_id"`
	Name      string    `json:"name,omitempty" db:"name"`
	Path      string    `json:"path,omitempty" db:"path"`
	CreatedAt time.Time `json:"cretedAt,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}

//GameName game名の構造体
type GameName struct {
	Name string `json:"name,omitempty" db:"name"`
}

//GameCheck gameのID,名前,パスの構造体
type GameCheck struct {
	GameID string `json:"gameId,omitempty" db:"game_id"`
	Name   string `json:"name,omitempty" db:"name"`
	Path   string `json:"path,omitempty" db:"path"`
}
