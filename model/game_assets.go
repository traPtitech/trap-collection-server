package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

var (
	// AssetTypeURL ゲームの本体の種類(URL)
	AssetTypeURL        uint8 = 0
	// AssetTypeJar ゲームの本体の種類(.jar)
	AssetTypeJar        uint8 = 1
	// AssetTypeWindowsExe ゲームの本体の種類(Windowsの.exe)
	AssetTypeWindowsExe uint8 = 2
	// AssetTypeMacApp ゲームの本体の種類(Macの.app)
	AssetTypeMacApp     uint8 = 3
)

// GameAsset gameのassetの構造体
type GameAsset struct {
	ID            uint `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	GameVersionID uint `gorm:"type:int(11);NOT NULL;"`
	GameVersion   GameVersion
	Type          uint8  `gorm:"type:tinyint;NOT NULL;"`
	Md5           string `gorm:"type:char(32);"`
	URL           string `gorm:"type:text"`
}

// GameAssetMeta game_assetsテーブルのリポジトリ
type GameAssetMeta interface {
	InsertGameURL(gameID string, url string) (*openapi.GameUrl, error)
}

// InsertGameURL ゲームのURLの追加
func (*DB) InsertGameURL(gameID string, url string) (*openapi.GameUrl, error) {
	var gameURL openapi.GameUrl
	err := db.Transaction(func(tx *gorm.DB) error {
		gameVersion := GameVersion{}
		err := tx.Where("game_id = ?", gameID).
			Select("id").
			First(&gameVersion).Error
		if err != nil {
			return fmt.Errorf("failed to get game version by game id: %w", err)
		}

		gameAsset := GameAsset{
			GameVersionID: gameVersion.ID,
			Type:          AssetTypeURL,
			URL:           url,
		}
		err = tx.Create(&gameAsset).Error
		if err != nil {
			return fmt.Errorf("failed to insert game asset: %w", err)
		}

		err = tx.Last(&gameAsset).Error
		if err != nil {
			return fmt.Errorf("failed to get the last game asset record: %w", err)
		}
		gameURL = openapi.GameUrl{
			Id:  int32(gameAsset.ID),
			Url: gameAsset.URL,
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return &gameURL, nil
}
