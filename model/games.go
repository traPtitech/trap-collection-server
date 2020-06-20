package model
//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"fmt"
	"log"
	"time"

	"github.com/traPtitech/trap-collection-server/openapi"
)

// Game gameの構造体
type Game struct {
	ID          string    `gorm:"type:varchar(36);PRIMARY_KEY;"`
	GameVersion GameVersion `gorm:"association_foreignkey:GameID;"`
	Name        string    `gorm:"type:varchar(32);NOT NULL;"`
	Description string    `gorm:"type:text;"`
	CreatedAt   time.Time `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameMeta gameテーブルのリポジトリ
type GameMeta interface {
	GetGameInfo(gameID string) (*openapi.Game, error)
}

// GetGameInfo ゲーム情報の取得
func (*DB) GetGameInfo(gameID string) (*openapi.Game, error) {
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
