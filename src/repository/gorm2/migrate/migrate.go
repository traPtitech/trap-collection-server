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
