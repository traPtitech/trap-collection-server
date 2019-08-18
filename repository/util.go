package repository

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

var (
	//Db db
	Db *sqlx.DB
)

//Establish データベースに接続
func Establish() error {
	_db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE")))
	if err != nil {
		return err
	}
	Db = _db

	return nil
}
