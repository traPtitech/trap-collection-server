package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

type AssetType string

const (
	// AssetTypeURL ゲームの本体の種類(URL)
	AssetTypeURL AssetType = "url"
	// AssetTypeJar ゲームの本体の種類(.jar)
	AssetTypeJar AssetType = "jar"
	// AssetTypeWindowsExe ゲームの本体の種類(Windowsの.exe)
	AssetTypeWindowsExe AssetType = "windows"
	// AssetTypeMacApp ゲームの本体の種類(Macの.app)
	AssetTypeMacApp AssetType = "mac"
)

// GameAsset gameのassetの構造体
type GameAsset struct {
	ID            uint   `gorm:"type:int(11) unsigned auto_increment;PRIMARY_KEY;"`
	GameVersionID string `gorm:"type:varchar(36);NOT NULL;"`
	GameVersion   GameVersion
	Type          AssetType `gorm:"type:enum('url','jar','windows','mac');NOT NULL;"`
	Md5           string    `gorm:"type:char(32);"`
	URL           string    `gorm:"type:text"`
}

// GameAssetMeta game_assetsテーブルのリポジトリ
type GameAssetMeta interface {
	InsertGameURL(gameID string, url string) (*openapi.GameUrl, error)
	InsertGameFile(gameID string, fileType AssetType, md5 string) (*openapi.GameFile, error)
	IsValidAssetType(fileType string) bool
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

// InsertGameFile ゲームのファイルの追加
func (*DB) InsertGameFile(gameID string, fileType AssetType, md5 string) (*openapi.GameFile, error) {
	var gameFile openapi.GameFile
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
			Type:          fileType,
			Md5:           md5,
		}
		err = tx.Create(&gameAsset).Error
		if err != nil {
			return fmt.Errorf("failed to insert game asset: %w", err)
		}

		err = tx.Last(&gameAsset).Error
		if err != nil {
			return fmt.Errorf("failed to get the last game asset record: %w", err)
		}
		gameFile = openapi.GameFile{
			Id:   int32(gameAsset.ID),
			Type: string(gameAsset.Type),
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed in transaction: %w", err)
	}

	return &gameFile, nil
}

func (*DB) IsValidAssetType(fileType string) bool {
	return fileType == string(AssetTypeJar) || fileType == string(AssetTypeWindowsExe) || fileType == string(AssetTypeMacApp)
}
