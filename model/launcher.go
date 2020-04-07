package model

import (
	"fmt"
	"log"
	"time"
)

// LauncherVersion ランチャーのバージョンの構造体
type LauncherVersion struct {
	ID uint `json:"id" gorm:"type:int(11) unsigned;PRIMARY_KEY;AUTO_INCREMENT;default:0;"`
	Name string `json:"name,omitempty" gorm:"type:varchar(32);NOT NULL;UNIQUE;"`
	GameVersionRelations []GameVersionRelation `json:"games" gorm:"foreignkey:LauncherVersionID;"`
	Questions []Question `json:"questions" gorm:"foreignkey:LauncherVersionID;"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"type:datetime;NOT NULL;default:CURRENT_TIMESTAMP;"`
	DeletedAt time.Time `json:"deleted_at,omitempty" gorm:"type:datetime;default:NULL;"`
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
	log.Println(key)
	isNotThere := db.Where("`key` = ?",key).First(&productKey).RecordNotFound()
	return !isNotThere
}

// GetLauncherVersionByID ランチャーのバージョンをIDから取得
func GetLauncherVersionByID(id uint) (LauncherVersion,error) {
	var launcherVersion LauncherVersion
	err := db.Where("id = ?",id).Preload("GameVersionRelations").Preload("Questions").First(&launcherVersion).Error
	if err != nil {
		return LauncherVersion{},fmt.Errorf("Failed In Getting Launcher Versions:%w",err)
	}
	return launcherVersion,nil
}