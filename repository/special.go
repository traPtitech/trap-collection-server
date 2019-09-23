package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

//InsertSpecial 特例の追加
func InsertSpecial(versionName string,gameName string,inout string) error {
	_,err := Db.Exec("Insert INTO special (id,version_name,game_name,inout) VALUES (?,?,?,?)", uuid.Must(uuid.NewV4()).String(), versionName, gameName, inout)
	if err != nil {
		return err
	}
	return nil
}

//DeleteSpecial 特例の削除
func DeleteSpecial(versionName string,gameName string) error {
	_,err := Db.Exec("UPDATE special deleted_at SET ? WHERE version_name=? AND game_name=? AND deleted_at IS NULL", time.Now(), versionName, gameName)
	if err != nil {
		return err
	}
	return nil
}

//DeleteSpecialByPeriod 特例の削除
func DeleteSpecialByPeriod(versionName string,startPeriod time.Time,endPeriod time.Time) error {
	_,err := Db.Exec("UPDATE special INNER JOIN game ON special.game_name=game.name special.deleted_at SET ? WHERE special.version_name=? AND game.time>? AND game.time<? AND special.deleted_at IS NULL", time.Now(), versionName, startPeriod, endPeriod)
	if err != nil {
		return err
	}
	return nil
}

//DeleteSpecialByVersion 特例の削除
func DeleteSpecialByVersion(versionName string) error {
	_,err := Db.Exec("UPDATE special deleted_at SET ? WHERE version_name=? AND deleted_at IS NULL", time.Now(), versionName)
	if err != nil {
		return err
	}
	return nil
}

//IsThereSpecial 同一の特例が存在するか
func IsThereSpecial(versionName string,gameName string) (bool,error) {
	var name string
	err := Db.Get(&name,"SELECT game_name FROM special WHERE version_name=? AND game_name=? AND deleted_at IS NULL", versionName, gameName)
	if err != nil {
		return false,err
	}
	return (name!=""),nil
}