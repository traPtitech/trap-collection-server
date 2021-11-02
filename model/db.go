package model

import (
	"fmt"
	"os"

	// sql init
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

var (
	db        *gorm.DB
	allTables = []interface{}{
		Game{},
		GameVersion{},
		GameAsset{},
		GameIntroduction{},
		LauncherVersion{},
		GameVersionRelation{},
		ProductKey{},
		AccessToken{},
		SeatVersion{},
		Seat{},
	}
)

// DBMeta DB関連のインターフェイス
type DBMeta interface {
	GameAssetMeta
	GameIntroductionMeta
	GameVersionRelationMeta
	GameVersionMeta
	GameMeta
	LauncherVersionMeta
	ProductKeyMeta
}

// DB DB関連の構造体
type DB struct{}

//EstablishDB データベースに接続
func EstablishDB() (*gorm.DB, error) {
	_db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOSTNAME"), os.Getenv("DB_PORT"), os.Getenv("DB_DATABASE"))+"?parseTime=true&loc=Asia%2FTokyo&charset=utf8mb4")
	if err != nil {
		return &gorm.DB{}, fmt.Errorf("Failed In Connecting To Databases:%w", err)
	}
	db = _db

	return db, nil
}

// Migrate DBのマイグレーション
func Migrate(env string) error {
	err := db.
		Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").
		AutoMigrate(allTables...).Error
	if err != nil {
		return fmt.Errorf("Failed In Migration:%w", err)
	}

	if env == "mock" || env == "development" {
		games := []Game{
			{
				ID:          "72c0c88c-27fd-4b58-b08e-e3307d2c17df",
				Name:        "ClayPlatesStory",
				Description: "ClayPlatesStory",
			},
			{
				ID:          "0900b29f-61db-478a-bc51-135f723daab1",
				Name:        "Darkray2",
				Description: "Darkray2",
			},
			{
				ID:          "813bb858-7c1f-4cfb-8d54-f1483634e390",
				Name:        "Flythm",
				Description: "Flythm",
			},
			{
				ID:          "269fd8b7-75f9-4029-b5b2-50f2a878f15c",
				Name:        "ガチャキング",
				Description: "ガチャキング",
			},
			{
				ID:          "b9ce327d-8ab8-4f4f-8fd1-714de175dc2a",
				Name:        "Intution",
				Description: "Intution",
			},
			{
				ID:          "b82b27c4-e837-497b-a099-4ccd08d19960",
				Name:        "TiteQuest",
				Description: "TiteQuest",
			},
		}
		for _, v := range games {
			err := db.FirstOrCreate(&v).Error
			if err != nil {
				return fmt.Errorf("Failed In Creating %s: %w", v.Name, err)
			}
		}

		versions := make([]*GameVersion, 0, len(games))
		for _, v := range games {
			versions = append(versions, &GameVersion{
				ID:          uuid.New().String(),
				GameID:      v.ID,
				Name:        "1",
				Description: v.Description,
			})
		}
		for _, v := range versions {
			err := db.Create(&v).Error
			if err != nil {
				return fmt.Errorf("Failed In Creating %s: %w", v.GameID, err)
			}
		}

		type gameDescription struct {
			gameType AssetType
			md5      string
			url      string
		}
		gameDescriptionMap := map[string]*gameDescription{
			"72c0c88c-27fd-4b58-b08e-e3307d2c17df": {
				gameType: AssetTypeWindowsExe,
				md5:      "9bf87a506c93f1511fca62a4a97fbb71",
			},
			"0900b29f-61db-478a-bc51-135f723daab1": {
				gameType: AssetTypeJar,
				md5:      "1e210c55cfb159002b18f69bad3677c6",
			},
			"813bb858-7c1f-4cfb-8d54-f1483634e390": {
				gameType: AssetTypeURL,
				url:      "https://flythm.trap.games/",
			},
			"269fd8b7-75f9-4029-b5b2-50f2a878f15c": {
				gameType: AssetTypeURL,
				url:      "https://gachaking.trap.games/",
			},
			"b9ce327d-8ab8-4f4f-8fd1-714de175dc2a": {
				gameType: AssetTypeWindowsExe,
				md5:      "f1b8fac5278fa2ecccaf9f2b7d927ab7",
			},
			"b82b27c4-e837-497b-a099-4ccd08d19960": {
				gameType: AssetTypeJar,
				md5:      "da0057ac91d58adde3836797fa2fd9ad",
			},
		}
		assets := make([]*GameAsset, 0, len(versions))
		for _, v := range versions {
			dsc := gameDescriptionMap[v.GameID]
			assets = append(assets, &GameAsset{
				ID:            uuid.New().String(),
				GameVersionID: v.ID,
				Type:          dsc.gameType,
				Md5:           dsc.md5,
				URL:           dsc.url,
			})
		}
		for _, v := range assets {
			err := db.Create(&v).Error
			if err != nil {
				return fmt.Errorf("Failed In Creating %d: %w", v.GameVersionID, err)
			}
		}

		introductions := make([]*GameIntroduction, 0, len(games)*2)
		for _, v := range games {
			introductions = append(introductions, &GameIntroduction{
				ID:        uuid.New().String(),
				GameID:    v.ID,
				Role:      0,
				Extension: 1,
			})
			introductions = append(introductions, &GameIntroduction{
				ID:        uuid.New().String(),
				GameID:    v.ID,
				Role:      1,
				Extension: 3,
			})
		}
		for _, v := range introductions {
			err := db.Create(&v).Error
			if err != nil {
				return fmt.Errorf("Failed In Creating %s: %w", v.GameID, err)
			}
		}

		db.Create(&ProductKey{
			Key:               os.Getenv("PRODUCT_KEY"),
			LauncherVersionID: "0831437a-e715-45d5-8d6a-0b4b065ae45d",
		})
	}

	return nil
}
