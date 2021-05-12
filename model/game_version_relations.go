package model

//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
	gormbulk "github.com/t-tiger/gorm-bulk-insert"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// GameVersionRelation ランチャーのバージョンに入るゲームの構造体
type GameVersionRelation struct {
	LauncherVersionID string `gorm:"type:varchar(36);NOT NULL;PRIMARY_KEY;"`
	LauncherVersion   LauncherVersion
	GameID            string `gorm:"type:varchar(36);NOT NULL;PRIMARY_KEY;"`
	Game              Game
}

// GameVersionRelationMeta game_version_relationテーブルのリポジトリ
type GameVersionRelationMeta interface {
	GetCheckList(versionID string, operatingSystem string) ([]*openapi.CheckItem, error)
	InsertGamesToLauncherVersion(launcherVersionID int, gameIDs []string) (*openapi.VersionDetails, error)
}

// GetCheckList チェックリストの取得
func (*DB) GetCheckList(versionID string, operatingSystem string) ([]*openapi.CheckItem, error) {
	types, ok := osGameTypeIntsMap[operatingSystem]
	if !ok {
		return []*openapi.CheckItem{}, errors.New("Unsupported OS")
	}
	query := db.Table("game_version_relations").
		Joins("INNER JOIN game_versions ON game_version_relations.game_id = game_versions.game_id").
		Joins("INNER JOIN game_assets ON game_versions.id = game_assets.game_version_id")
	rows, err := query.
		Joins("OUTER JOIN game_introductions ON").
		Select("game_version_relations.game_id, game_assets.md5, game_assets.type, game_assets.created_at, ?").
		Where("geme_assets.type IN ? AND game_version_relations.game_id IN ?",
			types,
			query.
				Select("game_version_relations.game_id, MAX(game_assets.created_at)").
				Where("game_version_relations.launcher_version_id = ? AND geme_assets.type IN ?", versionID, types).
				Group("game_version_relations.game_id").SubQuery()).Rows()
	if err != nil {
		return []*openapi.CheckItem{}, fmt.Errorf("Failed In Getting CheckList: %w", err)
	}

	var checkList []*openapi.CheckItem
	for rows.Next() {
		var checkItem *openapi.CheckItem
		err = rows.Scan(&checkItem.Id, &checkItem.Md5, &checkItem.Type, &checkItem.BodyUpdatedAt, &checkItem.ImgUpdatedAt, &checkItem.MovieUpdatedAt)
		if err != nil {
			return []*openapi.CheckItem{}, fmt.Errorf("Failed In Scanning CheckList: %w", err)
		}
		checkList = append(checkList, checkItem)
	}
	return checkList, nil
}

func (*DB) InsertGamesToLauncherVersion(launcherVersionID int, gameIDs []string) (*openapi.VersionDetails, error) {
	var version openapi.VersionDetails

	launcherVersion := LauncherVersion{}
	err := db.Where("id = ?", launcherVersionID).Find(&launcherVersion).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, errors.New("No Launcher Version")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get a launcher version by id: %w", err)
	}

	gameVersionRelations := make([]interface{}, 0, len(gameIDs))
	for _, gameID := range gameIDs {
		gameVersionRelation := GameVersionRelation{
			LauncherVersionID: uint(launcherVersionID),
			GameID:            gameID,
		}

		gameVersionRelations = append(gameVersionRelations, gameVersionRelation)
	}

	err = gormbulk.BulkInsert(db, gameVersionRelations, 3000)
	if err != nil {
		return nil, fmt.Errorf("failed to insert games into version: %w", err)
	}

	games, err := getGameVersion(db, launcherVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game by launcher version id: %w", err)
	}

	version = openapi.VersionDetails{
		Id:        int32(launcherVersion.ID),
		Name:      launcherVersion.Name,
		Games:     games,
		CreatedAt: launcherVersion.CreatedAt,
	}

	return &version, nil
}

func getGameVersion(db *gorm.DB, launcherVersionID int) ([]openapi.GameMeta, error) {
	//IDだけなのがなにか気持ち悪いので他のカラムも入れられるようPluckではなくSelectにしている
	rows, err := db.Table("game_version_relations").
		Joins("LEFT OUTER JOIN games ON game_version_relations.game_id = games.id").
		Where("game_version_relations.launcher_version_id = ?", launcherVersionID).
		Select("games.*").
		Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get games by launcher id: %w", err)
	}

	games := []openapi.GameMeta{}
	for rows.Next() {
		game := Game{}
		err := db.ScanRows(rows, &game)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		games = append(games, openapi.GameMeta{
			Id:   game.ID,
			Name: game.Name,
		})
	}

	return games, nil
}
