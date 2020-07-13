package model
//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// GameVersion gameのversionの構造体
type GameVersion struct {
	ID          uint      `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	GameID      string    `gorm:"type:varchar(36);NOT NULL;"`
	Name        string    `gorm:"type:varchar(36);NOT NULL;"`
	Description string    `gorm:"type:text;"`
	CreatedAt   time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameVersionMeta game_versionテーブルのリポジトリ
type GameVersionMeta interface {
	GetGameType(gameID string, operatingSystem string) (string, error)
	GetURL(gameID string) (string, error)
}

// GetGameType ゲームの種類の取得
func (*DB) GetGameType(gameID string, operatingSystem string) (string, error) {
	intOs, ok := osGameTypeIntMap[operatingSystem]
	if !ok {
		return "", errors.New("Invalid OS Error")
	}

	var intTypes []uint8
	err := db.Table("game_versions").
		Select("type").
		Joins("INNER JOIN game_assets ON game_versions.id = game_assets.game_version_id").
		Where("game_versions.game_id = ? AND game_assets.type IN (1,?)", gameID, intOs).
		Order("game_versions.created_at").
		Pluck("type", &intTypes).Error
	if err != nil {
		return "",fmt.Errorf("Failed In Getting Type: %w", err)
	}
	strType, ok := gameTypeIntStrMap[intTypes[0]]
	if !ok {
		log.Println("error: Unexpected Invalid Game Type")
		return "", errors.New("Invalid Game Type")
	}

	return strType, nil
}

// GetURL URLの取得
func (*DB) GetURL(gameID string) (string, error) {
	var url string
	rows, err := db.Table("game_versions").
		Select("game_assets.url").
		Joins("INNER JOIN game_assets ON game_versions.id = game_assets.game_version_id").
		Where("game_versions.game_id = ? AND game_assets.type = 0", gameID).
		Order("game_versions.created_at").
		Rows()
	if err != nil {
		return "", fmt.Errorf("Failed In Getting URL: %w", err)
	}
	if rows.Next() {
		err = rows.Scan(&url)
		if err != nil {
			return "", fmt.Errorf("Failed In Scaning Game URL: %w", err)
		}
	}

	return url, err
}
