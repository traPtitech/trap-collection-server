package repository

import (
	"time"

	"github.com/gofrs/uuid"

	"github.com/traPtitech/trap-collection-server/model"
)

//InsertSeat 席の挿入
func InsertSeat(x int, y int) error {
	_, err := Db.Exec("INSERT INTO seat (id,x,y,created_at) VALUES (?,?,?,?)", uuid.Must(uuid.NewV4()).String(), x, y, time.Now())
	if err != nil {
		return err
	}
	return nil
}

//DeleteSeat 席の削除
func DeleteSeat(x int, y int) error {
	_, err := Db.Exec("UPDATE seat SET daleted_at=? WHERE x=? AND y=? AND deleted_at IS NULL", time.Now(), x, y)
	if err != nil {
		return err
	}
	return nil
}

//GetSeat 埋まっている席の一覧
func GetSeat() ([]model.GetSeat, error) {
	seat := []model.GetSeat{}
	err := Db.Select(&seat, "SELECT x,y FROM seat WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	return seat, nil
}

//IsThereSeat 席の状態確認
func IsThereSeat(x int, y int) (bool, error) {
	var id string
	err := Db.Get(&id, "SELECT id FROM seat WHERE x=? AND y=? AND deleted_at IS NULL", x, y)
	if err != nil {
		return false, err
	}
	if id == "" {
		return false, nil
	}
	return true, nil
}
