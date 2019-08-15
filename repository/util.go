package repository

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/srinathgs/mysqlstore"
)

var (
	//Db db
	Db *sqlx.DB
	//Store sqlstore
	Store *mysqlstore.MySQLStore
)

//Establish データベースに接続
func Establish() (*mysqlstore.MySQLStore, error) {
	_db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE")))
	if err != nil {
		return nil, err
	}
	Db = _db

	Store, err := mysqlstore.NewMySQLStoreFromConnection(model.Db.DB, "sessions", "/", 60*60*24*14, []byte("secret-token"))
	if err != nil {
		return nil, err
	}

	return Store, nil
}
