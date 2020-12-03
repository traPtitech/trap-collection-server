package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
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
	IsExistGame(gameID string) (bool, error)
	GetGames(userID ...string) ([]*openapi.Game, error)
	PostGame(userID string, gameName string, description string) (*openapi.GameMeta, error)
	DeleteGame(gameID string) error
	GetGameInfo(gameID string) (*openapi.Game, error)
	UpdateGame(gameID string, gameMeta *openapi.NewGameMeta) (*openapi.GameMeta, error)
	CheckGameIDs(gameIDs []string) error
}

// IsExistGame ゲームが存在するかの確認
func (*DB) IsExistGame(gameID string) (bool, error) {
	err := db.Where("id = ?", gameID).
		Find(&Game{}).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get a game by id: %w", err)
	}

	return true, nil
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
			Where("(gv.id = ? OR gv.id IS NULL AND g.deleted_at IS NULL) AND maintainers.user_id = ?", sub, userID[0]).
			Rows()
	} else {
		rows, err = db.Where("gv.id = ? AND g.deleted_at IS NULL", sub).Rows()
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

// DeleteGame ゲームの削除
func (*DB) DeleteGame(gameID string) error {
	isNotFound := db.Where("id = ? AND deleted_at IS NULL", gameID).Find(&Game{}).RecordNotFound()
	if isNotFound {
		return errors.New("record not found")
	}

	err := db.Model(&Game{}).Where("id = ? AND deleted_at IS NULL", gameID).Update("deleted_at", time.Now()).Error
	if err != nil {
		return fmt.Errorf("failed to DELETE Game: %w", err)
	}

	return nil
}

// GetGameInfo ゲーム情報の取得
func (*DB) GetGameInfo(gameID string) (*openapi.Game, error) {
	game := &openapi.Game{
		Version: &openapi.GameVersion{},
	}
	rows, err := db.Table("games").
		Select("games.id, games.name, games.created_at, game_versions.id, game_versions.name, game_versions.description, game_versions.created_at").
		Joins("LEFT OUTER JOIN game_versions ON games.id = game_versions.game_id").
		Where("games.id = ?", gameID).
		Order("game_versions.created_at").
		Limit(1).
		Rows()
	if err != nil {
		return &openapi.Game{}, fmt.Errorf("Failed In Getting Game Info: %w", err)
	}
	if rows.Next() {
		var versionID sql.NullInt32
		var versionName sql.NullString
		var versionDescription sql.NullString
		var versionCreatedAt sql.NullTime
		err = rows.Scan(&game.Id, &game.Name, &game.CreatedAt, &versionID, &versionName, &versionDescription, &versionCreatedAt)
		if err != nil {
			return &openapi.Game{}, fmt.Errorf("Failed In Scaning Game Info: %w", err)
		}
		if versionID.Valid && versionName.Valid && versionDescription.Valid && versionCreatedAt.Valid {
			game.Version.Id = versionID.Int32
			game.Version.Name = versionName.String
			game.Version.Description = versionDescription.String
			game.Version.CreatedAt = versionCreatedAt.Time
		}
	}
	log.Printf("debug: %#v\n", game)

	return game, nil
}

// UpdateGame ゲームの更新
func (*DB) UpdateGame(gameID string, newGameMeta *openapi.NewGameMeta) (*openapi.GameMeta, error) {
	err := db.Model(&Game{}).Where("id = ? AND deleted_at IS NULL", gameID).Update(Game{
		Name:        newGameMeta.Name,
		Description: newGameMeta.Description,
	}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	var game Game
	err = db.Where("id = ?", gameID).Find(&game).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}

	gameMeta := &openapi.GameMeta{
		Id:          game.ID,
		Name:        game.Name,
		Description: game.Description,
		CreatedAt:   game.CreatedAt,
	}

	return gameMeta, nil
}

// InvalidGameIDs Checkをfailした理由ごとの不正なIDの配列
type InvalidGameIDs struct {
	NotFound  []string
	NoVersion []string
	NoAssets  []string
}

func (ids *InvalidGameIDs) Error() string {
	return fmt.Sprintf("invalid gameIds(not found:%s, no version: %s, no assets: %s)", strings.Join(ids.NotFound, "/"), strings.Join(ids.NoVersion, "/"), strings.Join(ids.NoAssets, "/"))
}

type gameIDState int

const (
	notFound gameIDState = iota
	noVersion
	noAsset
	ok
)

// CheckGameIDs gameIDが登録済みのものに該当するか確認
func (*DB) CheckGameIDs(gameIDs []string) error {
	rows, err := db.
		Where("games.id IN (?)", gameIDs).
		Table("games").
		Joins("LEFT OUTER JOIN game_versions ON games.id = game_versions.game_id").
		Joins("LEFT OUTER JOIN game_assets ON game_versions.id = game_assets.game_version_id").
		Group("games.id").
		Select("games.id, COUNT(game_versions.id), COUNT(game_assets.id)").
		Rows()
	if err != nil {
		return fmt.Errorf("failed to find gameIDs: %w", err)
	}

	gameIDMap := make(map[string]gameIDState, len(gameIDs))
	for _, gameID := range gameIDs {
		gameIDMap[gameID] = notFound
	}

	for rows.Next() {
		var gameID string
		var versionNum int
		var assetNum int
		err := rows.Scan(&gameID, &versionNum, &assetNum)
		if err != nil {
			return fmt.Errorf("failed to scan gameIDs: %w", err)
		}

		if versionNum == 0 {
			gameIDMap[gameID] = dontHaveVersion
			continue
		}
		if assetNum == 0 {
			gameIDMap[gameID] = dontHaveAsset
			continue
		}

		gameIDMap[gameID] = found
	}

	gameIDerr := &GameIDsError{}
	for gameID, state := range gameIDMap {
		switch state {
		case notFound:
			gameIDerr.NotFoundGameIDs = append(gameIDerr.NotFoundGameIDs, gameID)
		case dontHaveVersion:
			gameIDerr.DontHaveVersionGameIDs = append(gameIDerr.DontHaveVersionGameIDs, gameID)
		case dontHaveAsset:
			gameIDerr.DontHaveAssetGameIDs = append(gameIDerr.DontHaveAssetGameIDs, gameID)
		}
	}

	if len(gameIDerr.NotFoundGameIDs) == 0 && len(gameIDerr.DontHaveVersionGameIDs) == 0 && len(gameIDerr.DontHaveAssetGameIDs) == 0 {
		return nil
	}

	return gameIDerr
}
