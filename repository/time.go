package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

//InsertTime 時間の追加
func InsertTime(versionID string, gameID string, startTime time.Time, endTime time.Time) error {
	_, err := Db.Exec("INSERT INTO time (id,version_id,game_id,start_time,end_time) VALUES (?,?,?,?,?)", uuid.Must(uuid.NewV4()).String(), versionID, gameID, startTime, endTime)
	if err != nil {
		return err
	}
	return nil
}
