package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Game gameの構造体
type Game struct {
	ID          string    `gorm:"type:varchar(36);PRIMARY_KEY;"`
	Name        string    `gorm:"type:varchar(32);NOT NULL;"`
	Description string    `gorm:"type:text;"`
	CreatedAt   time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameVersion gameのversionの構造体
type GameVersion struct {
	ID          uint      `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameID      string    `gorm:"type:varchar(36);NOT NULL;"`
	Game        Game      `gorm:"FOREIGNKEY:GameID"`
	Name        string    `gorm:"type:varchar(36);NOT NULL;"`
	Description string    `gorm:"type:text;"`
	CreatedAt   time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameAsset gameのassetの構造体
type GameAsset struct {
	ID            uint `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameVersionID uint `gorm:"type:int(11);NOT NULL;"`
	GameVersion   GameVersion
	Type          uint8  `gorm:"type:tinyint;NOT NULL;"`
	Md5           string `gorm:"type:binary(16);"`
	URL           string `gorm:"type:text"`
}

// GameIntroduction gameのintroductionの構造体
type GameIntroduction struct {
	ID        uint   `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameID    string `gorm:"type:varchar(36);NOT NULL;"`
	Game      Game
	Role      uint8     `gorm:"type:tinyint;NOT NULL;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
}

// Maintainer gameのmaintainerの構造体
type Maintainer struct {
	ID        uint   `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;"`
	GameID    string `gorm:"type:varchar(36);NOT NULL;"`
	Game      Game
	UserID    string    `gorm:"type:varchar(32);NOT NULL;"`
	Role      uint8     `gorm:"type:tinyint;NOT NULL;DEFAULT:0;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt time.Time `gorm:"type:datetime;DEFAULT:NULL;"`
}

// CheckMaintainerID ゲームの管理者のチェック
func CheckMaintainerID(userID string, gameID string) (bool, error) {
	var maintainer Maintainer
	err := db.Select("user_id").Where("game_id = ? AND user_id = ?", gameID, userID).First(&maintainer).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
