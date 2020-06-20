package model

// GameAsset gameのassetの構造体
type GameAsset struct {
	ID            uint `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	GameVersionID uint `gorm:"type:int(11);NOT NULL;"`
	GameVersion   GameVersion
	Type          uint8  `gorm:"type:tinyint;NOT NULL;"`
	Md5           string `gorm:"type:char(32);"`
	URL           string `gorm:"type:text"`
}