package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

//InsertSeat 席の挿入
func InsertSeat(id string) error {
	_, err := Db.Exec("INSERT INTO seat (id,seat_id,created_at) VALUES (?,?,?)", uuid.Must(uuid.NewV4()).String(), id, time.Now())
	if err != nil {
		return err
	}
	return nil
}

//DeleteSeat 席の削除
func DeleteSeat(id string) error {
	_, err := Db.Exec("UPDATE seat SET deleted_at=? WHERE seat_id=? AND deleted_at IS NULL", time.Now(), id)
	if err != nil {
		return err
	}
	return nil
}

//GetSeat 埋まっている席の一覧
func GetSeat() ([]string, error) {
	var seat []string
	err := Db.Select(&seat, "SELECT seat_id FROM seat WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	return seat, nil
}

//IsThereSeat 席の状態確認
func IsThereSeat(seatID string) bool {
	var id string
	err := Db.Get(&id, "SELECT id FROM seat WHERE seat_id=? AND deleted_at IS NULL", seatID)
	if err != nil {
		return false
	}
	return true
}
