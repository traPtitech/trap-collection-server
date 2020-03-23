package model

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/jinzhu/gorm"
)

var (
	db     *gorm.DB
	client *gophercloud.ServiceClient
)

//EstablishDB データベースに接続
func EstablishDB() (*sql.DB, error) {
	_db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))+"?loc=Asia%2FTokyo&charset=utf8mb4")
	if err != nil {
		return &sql.DB{}, err
	}
	db = _db

	return db.DB(), nil
}

//EstablishConoHa ConoHaの認証
func EstablishConoHa() error {
	option, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return fmt.Errorf("Failed In Reading Auth Env:%w", err)
	}

	provider, err := openstack.AuthenticatedClient(option)
	if err != nil {
		return fmt.Errorf("Failed In Authorization:%w", err)
	}

	client, err = openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return fmt.Errorf("Failed In Reading Connecting To Storage:%w", err)
	}
	result := containers.Create(client, "trap_collection", nil)
	if result.Err != nil {
		return fmt.Errorf("Failed In Making New Storage:%w", err)
	}
	return nil
}
