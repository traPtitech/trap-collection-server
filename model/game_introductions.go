package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
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
	InsertIntroduction(gameID string, role string, ext string) error
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

// InsertIntroduction 画像・動画の追加
func (*DB) InsertIntroduction(gameID string, role string, ext string) error {
	intRole, ok := roleStrIntMap[role]
	if !ok {
		return errors.New("invalid role")
	}

	intExt, ok := extStrIntMap[ext]
	if !ok {
		return fmt.Errorf("invalid extension(%s)", ext)
	}

	introduction := &GameIntroduction{
		GameID:    gameID,
		Role:      intRole,
		Extension: intExt,
	}

	err := db.Create(introduction).Error
	if err != nil {
		return fmt.Errorf("failed to insert introduction: %w", err)
	}

	return nil
}
