package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
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
	GetPlayers() ([]int32, error)
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
	productKeyID, err := getKeyIDByKey(productKey)
	if err != nil {
		return fmt.Errorf("Failed In Getting KeyID: %w", err)
	}

	var player Player
	err = db.
		Where("product_key_id = ? AND ended_at IS NULL", productKeyID).
		First(&player).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return fmt.Errorf("failed to get a player: %w", err)
	}
	if !gorm.IsRecordNotFoundError(err) {
		return errors.New("Last Player Is Not End")
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

// GetPlayers プレイヤーを取得
func (*DB) GetPlayers() ([]int32, error) {
	players := []uint{}
	err := db.
		Model(&Player{}).
		Where("ended_at IS NULL").
		Pluck("product_key_id", &players).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	resPlayers := make([]int32, 0, len(players))
	for _, player := range players {
		resPlayers = append(resPlayers, int32(player))
	}

	return resPlayers, nil
}
