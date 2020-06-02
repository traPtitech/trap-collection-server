package model

import (
	"errors"
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

var gameTypeMap map[uint8]string = map[uint8]string{
	0: "url",
	1: "jar",
	2: "windows",
	3: "mac",
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
func GetGameInfo(gameID string) (*openapi.Game, error) {
	game := &openapi.Game{
		Version: &openapi.GameVersion{},
	}
	rows, err := db.Table("games").
		Select("games.id, games.name, games.created_at, game_versions.id, game_versions.name, game_versions.description, game_versions.created_at").
		Joins("INNER JOIN game_versions ON games.id = game_versions.game_id").
		Where("games.id = ?", gameID).
		Order("game_versions.created_at").
		Limit(1).
		Rows()
	if err != nil {
		return &openapi.Game{}, fmt.Errorf("Failed In Getting Game Info: %w", err)
	}
	if rows.Next() {
		err = rows.Scan(&game.Id, &game.Name, &game.CreatedAt, &game.Version.Id, &game.Version.Name, &game.Version.Description, &game.Version.CreatedAt)
		if err != nil {
			return &openapi.Game{}, fmt.Errorf("Failed In Scaning Game Info: %w", err)
		}
	}
	log.Printf("debug: %#v\n", game)

	return game, nil
}

var extMap = map[uint8]string{
	0: "jpeg",
	1: "png",
	2: "gif",
	3: "mp4",
}

// GetExtension 拡張子の取得
func GetExtension(gameID string, role int8) (string, error) {
	var gameIntroduction GameIntroduction
	err := db.Table("game_introductions").
		Select("extension").
		Where("game_id = ? AND role = ?", gameID, role).
		Order("created_at").
		First(&gameIntroduction).Error
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Extension: %w", err)
	}
	ext, ok := extMap[gameIntroduction.Extension]
	if !ok {
		log.Println("error: unexpected ext")
		return "", fmt.Errorf("Failed In ExtMap: %w", err)
	}
	return ext, nil
}

// GetURL URLの取得
func GetURL(gameID string) (string, error) {
	var url string
	rows, err := db.Table("game_versions").
		Select("game_assets.url").
		Joins("INNER JOIN game_assets ON game_versions.id = game_assets.game_version_id").
		Where("game_versions.game_id = ? AND game_assets.type = 0", gameID).
		Order("game_versions.created_at").
		Rows()
	if err != nil {
		return "", fmt.Errorf("Failed In Getting URL: %w", err)
	}
	if rows.Next() {
		err = rows.Scan(&url)
		if err != nil {
			return "", fmt.Errorf("Failed In Scaning Game URL: %w", err)
		}
	}

	return url, err
}

// GetGameType ゲームの種類の取得
func GetGameType(gameID string, operatingSystem string) (string, error) {
	osMap := map[string]uint{
		"windows": 2,
		"mac": 3,
	}
	intOs, ok := osMap[operatingSystem]
	if !ok {
		return "", errors.New("Invalid OS Error")
	}

	var intType uint8
	err := db.Table("game_versions").
		Select("type").
		Joins("INNER JOIN game_assets ON game_versions.id = game_assets.game_version_id").
		Where("game_versions.game_id = ? AND game_assets.type IN (1,?)", gameID, intOs).
		Order("game_versions.created_at").
		First(&intType).Error
	if err != nil {
		return "",fmt.Errorf("Failed In Getting Type: %w", err)
	}
	strType, ok := gameTypeMap[intType]
	if !ok {
		log.Println("error: Unexpected Invalid Game Type")
		return "", errors.New("Invalid Game Type")
	}

	return strType, nil
}
