package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

//InsertSpecial 特例の追加
func InsertSpecial(versionID string, gameName string, inout string) error {
	_, err := Db.Exec("Insert INTO special (id,version_id,game_name,status) VALUES (?,?,?,?)", uuid.Must(uuid.NewV4()).String(), versionID, gameName, inout)
	if err != nil {
		return err
	}
	return nil
}

//DeleteSpecial 特例の削除
func DeleteSpecial(versionID string, gameName string) error {
	_, err := Db.Exec("UPDATE special deleted_at SET ? WHERE version_id=? AND game_name=? AND deleted_at IS NULL", time.Now(), versionID, gameName)
	if err != nil {
		return err
	}
	return nil
}

//DeleteSpecialByPeriod 特例の削除
func DeleteSpecialByPeriod(versionID string, startPeriod time.Time, endPeriod time.Time) error {
	_, err := Db.Exec("UPDATE special INNER JOIN game ON special.game_name=game.name special.deleted_at SET ? WHERE special.version_id=? AND game.time>? AND game.time<? AND special.deleted_at IS NULL", time.Now(), versionID, startPeriod, endPeriod)
	if err != nil {
		return err
	}
	return nil
}

//DeleteSpecialByVersion 特例の削除
func DeleteSpecialByVersion(versionID string) error {
	_, err := Db.Exec("UPDATE special deleted_at SET ? WHERE version_id=? AND deleted_at IS NULL", time.Now(), versionID)
	if err != nil {
		return err
	}
	return nil
}

//IsThereSpecial 同一の特例が存在するか
func IsThereSpecial(versionID string, gameName string) bool {
	var name string
	err := Db.Get(&name, "SELECT game_name FROM special WHERE version_id=? AND game_name=? AND deleted_at IS NULL", versionID, gameName)
	if err != nil {
		return false
	}
	return true
}
