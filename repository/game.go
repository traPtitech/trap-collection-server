package repository

import (
	"errors"
	"time"

	"github.com/gofrs/uuid"

	"github.com/traPtitech/trap-collection-server/model"
)

//AddGame gameテーブルにgameを追加するメソッド
func AddGame(name string, path string) error {
	_, err = Db.Exec("INSERT INTO game (game_id,name,path,created_at,updated_at) VALUES (?,?,?,?,?)", uuid.Must(uuid.NewV4()).String(), name, path, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

//UpdateGame gameテーブルのupdated_atを更新するメソッド
func UpdateGame(name string) error {
	_, err = Db.Exec("UPDATE game SET upgated_at=? WHERE name=?", time.Now(), name)
	if err != nil {
		return err
	}

	return nil
}

//DeleteGame gameテーブルからgameを削除するメソッド
func DeleteGame(name string) error {
	var gameID string
	err := Db.Get(&gameID, "SELECT game_id from game WHERE name=?", name)
	if err != nil {
		return err
	}
	if gameID == "" {
		return errors.New("game not found")
	}

	_, err = Db.Exec("DELETE FROM game WHERE name=?", name)
	if err != nil {
		return err
	}

	return nil
}

//GetGameNameList game名一覧を取得するメソッド
func GetGameNameList() ([]model.GameName, error) {
	games := []model.GameName{}
	err := Db.Select(&games, "SELECT name from game")
	if err != nil {
		return games, err
	}

	return games, nil
}

//GameCheckList game_id,name,pathの一覧を取得するメソッド
func GameCheckList() ([]model.GameCheck, error) {
	games := []model.GameCheck{}
	err := Db.Select(&games, "SELECT game_id,name,path FROM game")
	if err != nil {
		return games, err
	}

	return games, nil
}

//LastUpdatedAt 最後に更新された時刻を確認するメソッド
func LastUpdatedAt() (time.Time, error) {
	var updatedAt time.Time
	err := Db.Get(&updatedAt, "SELECT updated_at FROM game ORDER BY updated_at DESC LIMIT 1")
	if err != nil {
		return updatedAt, err
	}

	return updatedAt, nil
}

//GetPath gameのパスを取得するメソッド
func GetPath(name string) (string, error) {
	var path string
	err := Db.Get(&path, "SELECT path FROM game WHERE name=?", name)
	if err != nil {
		return path, err
	}

	return path, nil
}

//IsThereGame 同名のgameが存在するか確認するメソッド
func IsThereGame(name string) (bool, error) {
	var gameID string
	err := Db.Get(&gameID, "SELECT game_id from game WHERE name=?", name)
	if err != nil {
		return nil, err
	}

	return !(gameID == ""), nil
}
