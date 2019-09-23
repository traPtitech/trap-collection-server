package repository

import (
	"errors"
	"time"

	"github.com/gofrs/uuid"

	"github.com/traPtitech/trap-collection-server/model"
)

//AddGame gameテーブルにgameを追加するメソッド
func AddGame(name string, container string, fileName string,md5 string) error {
	_, err := Db.Exec("INSERT INTO game (id,name,container,file_name,md5,created_at,updated_at) VALUES (?,?,?,?,?,?,?)", uuid.Must(uuid.NewV4()).String(), name, container, fileName, md5, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

//UpdateGame gameテーブルのupdated_atを更新するメソッド
func UpdateGame(name string) error {
	_, err := Db.Exec("UPDATE game SET upgated_at=? WHERE name=? AND deleted_at IS NULL", time.Now(), name)
	if err != nil {
		return err
	}

	return nil
}

//UpdateGameTime ゲームの起点時間を変更
func UpdateGameTime(name string,t time.Time) error {
	_, err := Db.Exec("UPDATE game SET time=? upgated_at=? WHERE name=? AND deleted_at IS NULL", t, time.Now(), name)
	if err != nil {
		return err
	}

	return nil
}

//DeleteGame gameテーブルからgameを削除するメソッド
func DeleteGame(name string) error {
	var gameID string
	err := Db.Get(&gameID, "SELECT id from game WHERE name=? AND deleted_at IS NULL", name)
	if err != nil {
		return err
	}
	if gameID == "" {
		return errors.New("game not found")
	}

	_, err = Db.Exec("UPDATE game deleted_at = ? WHERE name=? AND deleted_at IS NULL", time.Now(), name)
	if err != nil {
		return err
	}

	return nil
}

//GetGameNameList game名一覧を取得するメソッド
func GetGameNameList() ([]model.GameName, error) {
	games := []model.GameName{}
	err := Db.Select(&games, "SELECT name from game WHERE deleted_at IS NULL")
	if err != nil {
		return games, err
	}

	return games, nil
}

//GameCheckList id,name,container,file_nameの一覧を取得するメソッド
func GameCheckList() ([]model.GameCheck, error) {
	games := []model.GameCheck{}
	err := Db.Select(&games, "SELECT id,name,md5 FROM game WHERE deleted_at IS NULL")
	if err != nil {
		return games, err
	}

	return games, nil
}

//LastUpdatedAt 最後に更新された時刻を確認するメソッド
func LastUpdatedAt() (time.Time, error) {
	var updatedAt time.Time
	err := Db.Get(&updatedAt, "SELECT updated_at FROM game WHERE deleted_at IS NULL ORDER BY updated_at DESC LIMIT 1")
	if err != nil {
		return updatedAt, err
	}

	return updatedAt, nil
}

//GetContainerAndFileName gameのパスを取得するメソッド
func GetContainerAndFileName(name string) (string, string, error) {
	var file model.GameContainerAndFileName
	err := Db.Get(&file, "SELECT container,file_name FROM game WHERE name=? AND deleted_at IS NULL", name)
	if err != nil {
		return "", "", err
	}

	return file.Contaiter, file.FileName, nil
}

//GetContainerByName ゲーム名からコンテナを取得する関数
func GetContainerByName(name string) (string, error) {
	var container string
	err := Db.Get(&container, "SELECT container FROM game WHERE name=? AND deleted_at IS NULL", name)
	if err != nil {
		return "", err
	}

	return container, nil
}

//IsThereGame 同名のgameが存在するか確認するメソッド
func IsThereGame(name string) (bool, error) {
	var gameID string
	err := Db.Get(&gameID, "SELECT id from game WHERE name=? AND deleted_at IS NULL", name)
	if err != nil {
		return false, err
	}

	return !(gameID == ""), nil
}
