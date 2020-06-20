package model
//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Maintainer gameのmaintainerの構造体
type Maintainer struct {
	ID        uint   `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	GameID    string `gorm:"type:varchar(36);NOT NULL;"`
	Game      Game
	UserID    string    `gorm:"type:varchar(32);NOT NULL;"`
	Role      uint8     `gorm:"type:tinyint;NOT NULL;DEFAULT:0;"`
	MimeType  string    `gorm:"type:text;NOT NULL;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt time.Time `gorm:"type:datetime;DEFAULT:NULL;"`
}

// MaintainerMeta maintainerテーブルのリポジトリ
type MaintainerMeta interface {
	CheckMaintainerID(userID string, gameID string) (bool, error)
}

// CheckMaintainerID ゲームの管理者のチェック
func (*DB) CheckMaintainerID(userID string, gameID string) (bool, error) {
	var maintainer Maintainer
	err := db.Select("user_id").
		Where("game_id = ? AND user_id = ?", gameID, userID).
		First(&maintainer).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}