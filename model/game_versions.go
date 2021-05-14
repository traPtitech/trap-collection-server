package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/traPtitech/trap-collection-server/openapi"
)

// GameVersion gameのversionの構造体
type GameVersion struct {
	ID          string    `gorm:"type:varchar(36);PRIMARY_KEY;"`
	GameID      string    `gorm:"type:varchar(36);NOT NULL;"`
	Name        string    `gorm:"type:varchar(36);NOT NULL;"`
	Description string    `gorm:"type:text;"`
	CreatedAt   time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameVersionMeta game_versionテーブルのリポジトリ
type GameVersionMeta interface {
	GetGameType(gameID string, operatingSystem string) (string, error)
	GetGameVersions(gameID string) ([]*openapi.GameVersion, error)
	GetURL(gameID string) (string, error)
	InsertGameVersion(gameID string, name string, description string) (*openapi.GameVersion, error)
}

// GetGameType ゲームの種類の取得
func (*DB) GetGameType(gameID string, operatingSystem string) (string, error) {
	switch operatingSystem {
	case "win32":
		operatingSystem = "windows"
	case "darwin":
		operatingSystem = "mac"
	}

	intOs, ok := osGameTypeIntMap[operatingSystem]
	if !ok {
		return "", errors.New("Invalid OS Error")
	}

	var types []string
	err := db.Table("game_versions").
		Select("type").
		Joins("INNER JOIN game_assets ON game_versions.id = game_assets.game_version_id").
		Where("game_versions.game_id = ? AND game_assets.type IN (1,?)", gameID, intOs).
		Order("game_versions.created_at").
		Pluck("type", &types).Error
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Type: %w", err)
	}
	strType := types[0]
	if !ok {
		log.Println("error: Unexpected Invalid Game Type")
		return "", errors.New("Invalid Game Type")
	}

	return strType, nil
}

// GetURL URLの取得
func (*DB) GetURL(gameID string) (string, error) {
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

// InsertGameVersion GameVersionの追加
func (*DB) InsertGameVersion(gameID string, name string, description string) (*openapi.GameVersion, error) {
	newGameVersion := GameVersion{
		GameID:      gameID,
		Name:        name,
		Description: description,
	}

	err := db.Create(&newGameVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create a game version record: %w", err)
	}

	gameVersion := GameVersion{}
	err = db.Last(&gameVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get a last game version record: %w", err)
	}

	apiGameVersion := openapi.GameVersion{
		Id:          int32(gameVersion.ID),
		Name:        gameVersion.Name,
		Description: gameVersion.Description,
		CreatedAt:   gameVersion.CreatedAt,
	}

	return &apiGameVersion, nil
}

// GetGameVersions ゲームのバージョンの取得
func (*DB) GetGameVersions(gameID string) ([]*openapi.GameVersion, error) {
	var gameVersions []*GameVersion
	err := db.Where("game_id = ?", gameID).Find(&gameVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get game versions: %w", err)
	}

	apiGameVersions := make([]*openapi.GameVersion, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		apiGameVersion := openapi.GameVersion{
			Id:          int32(gameVersion.ID),
			Name:        gameVersion.Name,
			Description: gameVersion.Description,
			CreatedAt:   gameVersion.CreatedAt,
		}

		apiGameVersions = append(apiGameVersions, &apiGameVersion)
	}

	return apiGameVersions, nil
}
