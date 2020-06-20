package model
//go:generate mockgen -source=$GOFILE -destination=mock_${GOFILE} -package=$GOPACKAGE

import (
	"errors"
	"fmt"

	"github.com/traPtitech/trap-collection-server/openapi"
)

// GameVersionRelation ランチャーのバージョンに入るゲームの構造体
type GameVersionRelation struct {
	LauncherVersionID uint `gorm:"type:int(11) auto_increment;NOT NULL;PRIMARY_KEY;"`
	LauncherVersion   LauncherVersion
	GameID            string `gorm:"type:varchar(36);NOT NULL;PRIMARY_KEY;"`
	Game              Game
}

// GameVersionRelationMeta game_version_relationテーブルのリポジトリ
type GameVersionRelationMeta interface {
	GetCheckList(versionID uint, operatingSystem string) ([]*openapi.CheckItem, error)
}

// GetCheckList チェックリストの取得
func (*DB) GetCheckList(versionID uint, operatingSystem string) ([]*openapi.CheckItem, error) {
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
