package model

import (
	"errors"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const containerName = "trap_collection"

var (
	db        *gorm.DB
	allTables = []interface{}{
		Game{},
		GameVersion{},
		GameAsset{},
		GameIntroduction{},
		Maintainer{},
		LauncherVersion{},
		ProductKey{},
		GameVersionRelation{},
		Player{},
		Question{},
		QuestionOption{},
		Response{},
		TextAnswer{},
		OptionAnswer{},
		GameRating{},
	}
)

//EstablishDB データベースに接続
func EstablishDB(parseTime bool) (*gorm.DB, error) {
	var str string
	if parseTime {
		str = "?parseTime=true&loc=Asia%2FTokyo&charset=utf8mb4"
	} else {
		str = "?loc=Asia%2FTokyo&charset=utf8mb4"
	}
	_db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))+str)
	if err != nil {
		return &gorm.DB{}, fmt.Errorf("Failed In Connecting To Databases:%w", err)
	}
	db = _db

	return db, nil
}

// Migrate DBのマイグレーション
func Migrate(env string) error {
	err := db.AutoMigrate(allTables...).Error
	if err != nil {
		return fmt.Errorf("Failed In Migration:%w", err)
	}

	if env == "test" {
		launcherVersion := LauncherVersion{Name: "dev"}
		err = db.Where("name=\"dev\"").FirstOrCreate(&launcherVersion).Error
		if err != nil {
			return fmt.Errorf("Failed In Select Or Creating A Dev Version:%w", err)
		}
		key := os.Getenv("PRODUCT_KEY")
		if len(key) == 0 {
			return errors.New("NO PRODUCT_KEY")
		}
		productKey := ProductKey{Key: key, LauncherVersionID: launcherVersion.ID}
		err = db.Where(productKey).FirstOrCreate(&productKey).Error
		if err != nil {
			return fmt.Errorf("Failed In Select Or Creating A Product Key:%w", err)
		}
	}

	return nil
}
