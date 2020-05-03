package model

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
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
	MimeType  string    `gorm:"type:text;NOT NULL;"`
	CreatedAt time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt time.Time `gorm:"type:datetime;DEFAULT:NULL;"`
}

// CheckMaintainerID ゲームの管理者のチェック
func CheckMaintainerID(userID string, gameID string) (bool, error) {
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

// GetGameInfo ゲーム情報の取得
func GetGameInfo(gameID string) (openapi.Game, error) {
	type gameInfo struct {
		id string `db:"games.id"`
		name string `db:"games.name"`
		createdAt time.Time `db:"games.created_at"`
		versionID int32 `db:"launcher_versions.id"`
		versionName string `db:"launcher_versions.name"`
		versionDescription string `db:"launcher_versions.description"`
		versionCreatedAt time.Time `db:"launcher_versions.created_at"`
	}
	var game gameInfo
	err := db.Table("games").
		Select("games.id, games.name, games.created_at, launcher_versions.id, launcher_versions.name, launcher_versions.description, launcher_versions.created_at").
		Joins("INNER JOIN game_versions ON games.id = launcher_versions.game_id").
		Where("game.id = ?", gameID).
		Order("version.created_at").
		Scan(&game).Error
	if err != nil {
		return openapi.Game{}, fmt.Errorf("Failed In Getting Game Info: %w", err)
	}
	version := openapi.GameVersion{
		Id: game.versionID,
		Name: game.versionName,
		Description: game.versionDescription,
		CreatedAt: game.versionCreatedAt,
	}
	openapiGame := openapi.Game{
		Id: game.id,
		Name: game.name,
		CreatedAt: game.createdAt,
		Version: &version,
	}
	return openapiGame, nil
}

// GetExtension 拡張子の取得
func GetExtension(gameID string, role int8) (string, error) {
	var ext string
	err := db.Table("game_introductions").
		Select("extension").
		Where("game_id = ? AND role = ?", gameID, role).
		Order("created_at").
		First(&ext).Error
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Extension: %w", err)
	}
	return ext, nil
}

// GetURL URLの取得
func GetURL(gameID string) (string, error) {
	var url string
	err := db.Table("game_versions").
		Joins("INNER JOIN game_assets ON game_versions.id = game_assets.game_version_id").
		Where("game_versions.game_id = ? AND game_assets.type = 0", gameID).
		Order("game_versions.created_at").
		First(&url).Error
	if err != nil {
		return "", fmt.Errorf("Failed In Getting URL: %w", err)
	}
	return url, err
}
