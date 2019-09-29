package repository

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
)

var (
	//Db db
	Db     *sqlx.DB
	client *gophercloud.ServiceClient
)

//EstablishDB データベースに接続
func EstablishDB() error {
	_db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))+"?parseTime=true&loc=Asia%2FTokyo&charset=utf8mb4")
	if err != nil {
		return err
	}
	Db = _db

	return nil
}

//EstablishConoHa ConoHaの認証
func EstablishConoHa() error {
	_, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}

	option, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return err
	}

	provider, err := openstack.AuthenticatedClient(option)
	if err != nil {
		return err
	}

	client, err = openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return err
	}
	result := containers.Create(client, "game0", nil)
	if result.Err != nil {
		return result.Err
	}
	fmt.Println(result)
	return nil
}

//NullTimeToString 変換
func NullTimeToString(t mysql.NullTime) string {
	if t.Valid {
		return t.Time.Format(time.RFC3339)
	}
	return "NULL"
}

//NullStringConvert 変換
func NullStringConvert(str sql.NullString) string {
	if str.Valid {
		return str.String
	}
	return "NULL"
}

//GetUserID ユーザーIDの取得
func GetUserID(c echo.Context) string {
	res := c.Request().Header.Get("X-Showcase-User")
	// test用
	if res == "" {
		return "mds_boy"
	}
	return res
}
