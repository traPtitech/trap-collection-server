package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
	"fmt"
	"time"
)

// Player プレイヤーの履歴の構造体
type Player struct {
	ID           uint      `gorm:"type:int(11) unsigned auto_increment;NOT NULL;PRIMARY_KEY;"`
	ProductKeyID uint      `gorm:"type:int(11) unsigned;not null;"`
	StartedAt    time.Time `gorm:"type:datetime;not null;default:current_timestamp;"`
	EndedAt      time.Time `gorm:"type:datetime;default:null;"`
}

// PlayerMeta playerテーブルのリポジトリ
type PlayerMeta interface {
	PostPlayer(productKey string) error
	DeletePlayer(productKey string) error
}

func getPlayerIDByProductKey(productKey string) (playerID uint, err error) {
	err = db.Where("product_key = ?", productKey).Select("id").Find(&playerID).Error
	if err != nil {
		return 0, fmt.Errorf("Failed In Getting PlayerID: %w", err)
	}
	return playerID, nil
}

// PostPlayer プレイヤーの追加
func (*DB) PostPlayer(productKey string) error {
	var player Player
	isNotTherePlayer := db.Where("product_key_id = ? AND ended_at IS NULL", productKey).
		First(&player).
		RecordNotFound()
	if !isNotTherePlayer {
		return errors.New("Last Player Is Not End")
	}

	productKeyID, err := getKeyIDByKey(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Getting KeyID: %w", err)
	}

	player = Player{
		ProductKeyID: productKeyID,
	}

	err = db.Create(&player).Error
	if err != nil {
		return fmt.Errorf("Failed In Creating Player: %w", err)
	}

	return nil
}

// DeletePlayer プレイヤーを削除
func (*DB) DeletePlayer(productKey string) error {
	productKeyID, err := getKeyIDByKey(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Getting KeyID: %w", err)
	}

	err = db.Table("players").
		Where("product_key_id = ? AND ended_at IS NULL", productKeyID).
		Update("ended_at", time.Now()).
		Error
	if err != nil {
		return fmt.Errorf("Failed In Updating EndedAt: %w", err)
	}

	return nil
}
