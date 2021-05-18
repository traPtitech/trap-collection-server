package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"database/sql"
	"fmt"
	"time"
)

// ProductKey ProductKeyの構造体
type ProductKey struct {
	ID                string       `gorm:"type:varchar(29);NOT NULL;PRIMARY_KEY;"`
	Key               string       `gorm:"type:varchar(29);NOT NULL;PRIMARY_KEY;default:\"\";"`
	LauncherVersionID uint         `gorm:"type:int(11) unsigned;"`
	CreatedAt         time.Time    `gorm:"type:datetime;NOT NULL;"`
	DeletedAt         sql.NullTime `gorm:"type:datetime;DEFAULT:NULL;"`
}

// ProductKeyMeta product_keyテーブルのリポジトリ
type ProductKeyMeta interface {
	CheckProductKey(key string) (bool, uint)
}

func getKeyIDByKey(key string) (uint, error) {
	productKey := ProductKey{}
	err := db.Where("`key` = ?", key).First(&productKey).Error
	if err != nil {
		return 0, fmt.Errorf("Failed In Getting Key ID: %w", err)
	}
	return productKey.ID, nil
}

// CheckProductKey プロダクトキーが正しいか確認
func (*DB) CheckProductKey(key string) (bool, uint) {
	productKey := ProductKey{}
	isNotThere := db.Where("`key` = ?", key).First(&productKey).RecordNotFound()
	return !isNotThere, productKey.LauncherVersionID
}
