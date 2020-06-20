package model
//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"
	"log"
	"time"
)

// GameIntroduction gameのintroductionの構造体
type GameIntroduction struct {
	ID        uint   `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	GameID    string `gorm:"type:varchar(36);NOT NULL;"`
	Game      Game
	Role      uint8     `gorm:"type:tinyint;NOT NULL;"`
	Extension uint8     `gorm:"type:tinyint;NOT NULL;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
}

// GameIntroductionMeta game_introductionテーブルのリポジトリ
type GameIntroductionMeta interface {
	GetExtension(gameID string, role int8) (string, error)
}

// GetExtension 拡張子の取得
func (*DB) GetExtension(gameID string, role int8) (string, error) {
	var gameIntroduction GameIntroduction
	err := db.Table("game_introductions").
		Select("extension").
		Where("game_id = ? AND role = ?", gameID, role).
		Order("created_at").
		First(&gameIntroduction).Error
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Extension: %w", err)
	}
	ext, ok := extIntStrMap[gameIntroduction.Extension]
	if !ok {
		log.Println("error: unexpected ext")
		return "", fmt.Errorf("Failed In ExtMap: %w", err)
	}
	return ext, nil
}