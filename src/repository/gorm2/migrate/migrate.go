package migrate

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var (
	migrations = []*gormigrate.Migration{
		v1(), // アプリケーションのv1時へのマイグレーション
		v2(), // アプリケーションのv2用テーブルの追加
	}
)

func Migrate(db *gorm.DB, featureV2 bool) error {
	m := gormigrate.New(db.Session(&gorm.Session{}), &gormigrate.Options{
		TableName:                 "migrations",
		IDColumnName:              "id",
		IDColumnSize:              190,
		UseTransaction:            false,
		ValidateUnknownMigrations: true,
	}, migrations)

	m.InitSchema(func(db *gorm.DB) error {
		err := db.AutoMigrate(v1Tables...)
		if err != nil {
			return fmt.Errorf("failed to init schema: %w", err)
		}

		err = setupGameFileTypeTable(db)
		if err != nil {
			return fmt.Errorf("failed to setup game file type table: %w", err)
		}

		err = setupGameImageTypeTable(db)
		if err != nil {
			return fmt.Errorf("failed to setup game image type table: %w", err)
		}

		err = setupGameVideoTypeTable(db)
		if err != nil {
			return fmt.Errorf("failed to setup game video type table: %w", err)
		}

		err = setupGameManagementRoleTypeTable(db)
		if err != nil {
			return fmt.Errorf("failed to setup game management role type table: %w", err)
		}

		return nil
	})

	if featureV2 {
		err := m.Migrate()
		if err != nil {
			return fmt.Errorf("failed to migrate: %w", err)
		}
	} else {
		err := m.MigrateTo("1")
		if err != nil {
			return fmt.Errorf("failed to migrate to v1: %w", err)
		}
	}

	return nil
}
