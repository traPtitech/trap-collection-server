package model

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

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
		versionID int32 `db:"game_versions.id"`
		versionName string `db:"game_versions.name"`
		versionDescription string `db:"game_versions.description"`
		versionCreatedAt time.Time `db:"game_versions.created_at"`
	}
	var game gameInfo
	err := db.Table("games").
		Select("games.id, games.name, games.created_at, game_versions.id, game_versions.name, game_versions.description, game_versions.created_at").
		Joins("INNER JOIN game_versions ON games.id = game_versions.game_id").
		Where("games.id = ?", gameID).
		Order("game_versions.created_at").
		Limit(1).
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
