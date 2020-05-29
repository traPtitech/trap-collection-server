package model

import (
	"errors"
	"fmt"
	"time"
)

func getPlayerIDByProductKey(productKey string) (playerID uint, err error) {
	err = db.Where("product_key = ?", productKey).Select("id").Find(&playerID).Error
	if err != nil {
		return 0, fmt.Errorf("Failed In Getting PlayerID: %w", err)
	}
	return playerID, nil
}

// PostPlayer プレイヤーの追加
func PostPlayer(productKey string) error {
	var player Player
	isNotTherePlayer := db.Where("productKey = ? AND ended_at IS NULL", productKey).First(&player).RecordNotFound()
	if !isNotTherePlayer {
		return errors.New("Last Player Is Not End")
	}
	productKeyID,err := getKeyIDByKey(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Getting KeyID: %w", err)
	}
	player = Player{
		ProductKeyID: productKeyID,
	}
	err = db.Create(player).Error
	if err != nil {
		return fmt.Errorf("Failed In Creating Player: %w", err)
	}
	return nil
}

// DeletePlayer プレイヤーを削除
func DeletePlayer(productKey string) error {
	err := db.Where("productKey = ? AND ended_at IS NULL", productKey).Update("ended_at", time.Now).Error
	if err != nil {
		return fmt.Errorf("Failed In Updating EndedAt: %w", err)
	}
	return nil
}
