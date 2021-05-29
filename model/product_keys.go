package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"database/sql"
	"time"
)

// ProductKey ProductKeyの構造体
type ProductKey struct {
	ID                string       `gorm:"type:varchar(29);NOT NULL;PRIMARY_KEY;"`
	Key               string       `gorm:"type:varchar(29);NOT NULL;PRIMARY_KEY;default:\"\";"`
	LauncherVersionID string       `gorm:"type:varchar(36);"`
	CreatedAt         time.Time    `gorm:"type:datetime;NOT NULL;"`
	DeletedAt         sql.NullTime `gorm:"type:datetime;DEFAULT:NULL;"`
}

// ProductKeyMeta product_keyテーブルのリポジトリ
type ProductKeyMeta interface {
	CheckProductKey(key string) (bool, string)
}

// CheckProductKey プロダクトキーが正しいか確認
func (*DB) CheckProductKey(key string) (bool, string) {
	productKey := ProductKey{}
	isNotThere := db.Where("`key` = ?", key).First(&productKey).RecordNotFound()
	return !isNotThere, productKey.LauncherVersionID
}
