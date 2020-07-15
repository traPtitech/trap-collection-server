package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Game gameの構造体
type Game struct {
	ID          string      `gorm:"type:varchar(36);PRIMARY_KEY;"`
	GameVersion GameVersion `gorm:"association_foreignkey:GameID;"`
	Name        string      `gorm:"type:varchar(32);NOT NULL;"`
	Description string      `gorm:"type:text;"`
	CreatedAt   time.Time   `gorm:"type:datetime;NOT NULL;DEFAULT:CURRENT_TIMESTAMP;"`
	DeletedAt   time.Time   `gorm:"type:varchar(32);DEFAULT:NULL;"`
}

// GameMeta gameテーブルのリポジトリ
type GameMeta interface {
	GetGames(userID ...string) ([]*openapi.Game, error)
	PostGame(userID string, gameName string, description string) (*openapi.GameMeta, error)
	GetGameInfo(gameID string) (*openapi.Game, error)
}

// GetGames ゲーム一覧の取得
func (*DB) GetGames(userID ...string) ([]*openapi.Game, error) {
	sub := db.Table("games AS gs").
		Select("gvs.id").
		Joins("LEFT OUTER JOIN game_versions AS gvs ON gs.id = gvs.game_id").
		Where("gs.id = g.id").
		Order("gvs.created_at DESC").
		Limit(1).
		SubQuery()
	db := db.Table("games AS g").
		Select("g.id, g.name, g.created_at, gv.id, gv.name, gv.description, gv.created_at").
		Joins("LEFT OUTER JOIN game_versions AS gv ON g.id = gv.game_id")

	var rows *sql.Rows
	var err error
	if len(userID) != 0 {
		rows, err = db.Joins("INNER JOIN maintainers ON g.id = maintainers.game_id").
			Where("(gv.id = ? OR gv.id IS NULL) AND maintainers.user_id = ?", sub, userID[0]).
			Rows()
	} else {
		rows, err = db.Where("gv.id = ?", sub).Rows()
	}
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Games: %w", err)
	}

	var games []*openapi.Game
	for rows.Next() {
		game := &openapi.Game{}
		var id sql.NullInt32
		var name sql.NullString
		var description sql.NullString
		var createdAt sql.NullTime
		err = rows.Scan(&game.Id, &game.Name, &game.CreatedAt, &id, &name, &description, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("Failed In Scanning Game: %w", err)
		}
		if id.Valid && name.Valid && description.Valid && createdAt.Valid {
			game.Version = &openapi.GameVersion{
				Id:          id.Int32,
				Name:        name.String,
				Description: description.String,
				CreatedAt:   createdAt.Time,
			}
		}
		games = append(games, game)
	}

	return games, nil
}

// PostGame ゲームの追加
func (*DB) PostGame(userID string, gameName string, gameDescription string) (*openapi.GameMeta, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	game := &Game{
		ID:          id.String(),
		Name:        gameName,
		Description: gameDescription,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&game).Error
		if err != nil {
			return fmt.Errorf("failed to INSERT game record: %w", err)
		}

		err = tx.Last(&game).Error
		if err != nil {
			return fmt.Errorf("failed to GET added game record: %w", err)
		}

		maintainer := Maintainer{
			GameID: game.ID,
			UserID: userID,
			Role:   1,
		}
		err = tx.Create(&maintainer).Error
		if err != nil {
			return fmt.Errorf("failed to INSERT maintainer record: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Trasaction Error: %w", err)
	}

	gameMeta := &openapi.GameMeta{
		Id:          game.ID,
		Name:        game.Name,
		Description: game.Description,
		CreatedAt:   game.CreatedAt,
	}

	return gameMeta, nil
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
