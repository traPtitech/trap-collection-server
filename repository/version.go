package repository

import (
	"time"

	"github.com/gofrs/uuid"

	"github.com/traPtitech/trap-collection-server/model"
)

//InsertVersionForSale 販売用のバージョン追加
func InsertVersionForSale(name string,startPeriod time.Time,endPeriod time.Time,startTime time.Time) error {
	_,err := Db.Exec("INSERT INTO version_for_sale (id,name,start_period,end_period,start_time,created_at) VALUES (?,?,?,?,?,?)", uuid.Must(uuid.NewV4()).String(),name,startPeriod,endPeriod,startTime,time.Now())
	if err != nil {
		return err
	}

	return nil
}

//InsertVersionNotForSale 工大祭用のバージョン追加
func InsertVersionNotForSale(name string,questionnaireID int,startPeriod time.Time,endPeriod time.Time,startTime time.Time) error {
	_,err := Db.Exec("INSERT INTO version_not_for_sale (id,name,questionnaire_id,start_period,end_period,start_time,created_at) VALUES (?,?,?,?,?,?,?)", uuid.Must(uuid.NewV4()).String(),name,questionnaireID,startPeriod,endPeriod,startTime,time.Now())
	if err != nil {
		return err
	}

	return nil
}

//UpdateVersionForSale 販売用バージョンの変更
func UpdateVersionForSale(name string,startPeriod time.Time,endPeriod time.Time,startTime time.Time) error {
	_,err := Db.Exec("UPDATE version_for_sale (start_period,end_period,start_time,updated_at) SET (?,?,?,?) WHERE name = ? AND deleted_at IS NULL", startPeriod,endPeriod,startTime,time.Now(),name)
	if err != nil {
		return err
	}

	return nil
}

//UpdateVersionNotForSale 工大祭用のバージョンの変更
func UpdateVersionNotForSale(name string,questionnaireID int,startPeriod time.Time,endPeriod time.Time,startTime time.Time) error {
	_,err := Db.Exec("UPDATE version_for_sale (questionnaire_id,start_period,end_period,start_time,updated_at) SET (?,?,?,?,?) WHERE name = ? AND deleted_at IS NULL", questionnaireID, startPeriod,endPeriod,startTime,time.Now(),name)
	if err != nil {
		return err
	}

	return nil
}

//DeleteVersionForSale 販売用のバージョン削除
func DeleteVersionForSale(name string) error {
	_,err := Db.Exec("UPDATE version_for_sale SET deleted_at = ? WHERE name = ? AND deleted_at IS NULL",time.Now(),name)
	if err != nil {
		return err
	}

	return nil
}

//DeleteVersionNotForSale 工大祭用のバージョン削除
func DeleteVersionNotForSale(name string) error {
	_,err := Db.Exec("UPDATE version_not_for_sale SET deleted_at = ? WHERE name = ? AND deleted_at IS NULL",time.Now(),name)
	if err != nil {
		return err
	}

	return nil
}

//VersionForSaleList 販売用のバージョンの一覧
func VersionForSaleList() ([]model.VersionForSale,error) {
	versionList := []model.VersionForSale{}
	err := Db.Select(&versionList,"SELECT name,start_period,end_period,start_time FROM versions_for_sale WHERE deleted_at IS NULL")
	if err != nil {
		return nil,err
	}

	return versionList,nil
}

//VersionNotForSaleList 販売用のバージョンの一覧
func VersionNotForSaleList() ([]model.VersionNotForSale,error) {
	versionList := []model.VersionNotForSale{}
	err := Db.Select(&versionList,"SELECT name,questionnaire_id,start_period,end_period,start_time FROM versions_for_sale WHERE deleted_at IS NULL")
	if err != nil {
		return nil,err
	}

	return versionList,nil
}

//IsThereVersion 販売・工大祭用バージョンに同名のものが存在するか
func IsThereVersion(name string) (bool,error) {
	var versionForSaleName string
	err := Db.Get(&versionForSaleName,"SELECT name FROM version_for_sale WHERE name = ?",name)
	if err!=nil{
		return false,err
	}
	var versionNotForSaleName string
	err = Db.Get(&versionNotForSaleName,"SELECT name FROM version_not_for_sale WHERE name = ?",name)
	if err!=nil{
		return false,err
	}
	return ((versionForSaleName!="") && (versionNotForSaleName!="")),nil
}
