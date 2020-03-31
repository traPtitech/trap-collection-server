package model

import (
	"fmt"
	"time"
)

// LauncherVersion ランチャーのバージョンの構造体
type LauncherVersion struct {
	ID uint `gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;default:0;"`
	Name string `gorm:"type:varchar(32);NOT NULL;UNIQUE;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
	DeletedAt time.Time `gorm:"type:datetime;default:NULL;"`
}

// ProductKey アクセストークンの構造体
type ProductKey struct {
	Key string `gorm:"type:varchar(29);NOT NULL;PRIMARY_KEY;default:\"\";"`
	LauncherVersionID uint `gorm:"type:int(11) unsigned;"`
	LauncherVersion LauncherVersion
}

// CheckProductKey プロダクトキーが正しいか確認
func CheckProductKey(key string) bool {
	productKey := ProductKey{}
	fmt.Println(key)
	isThere := db.Where("`key` = ?",key).First(&productKey).RecordNotFound()
	return !isThere
}