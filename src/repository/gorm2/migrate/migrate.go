// migrate
//
// Deprecated: migrate パッケージは、古いマイグレーションのコードです。
// 代わりに [github.com/traPtitech/trap-collection-server/src/repository/gorm2/schema] パッケージを使用してください。
package migrate

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var (
	migrations = []*gormigrate.Migration{
		v1(),  // アプリケーションのv1時へのマイグレーション
		v2(),  // アプリケーションのv2用テーブルの追加
		v3(),  // v2でmigrationし忘れていたgameTable2のmigration
		v4(),  // v2_game_versionsにurlカラムを追加
		v5(),  // v2_game_filesのunique制約を解除
		v6(),  // v2_product_keysのstatusカラムを追加
		v7(),  // adminテーブルの追加
		v8(),  // v2 api用のデータのマイグレーション
		v9(),  // seatテーブルの追加
		v10(), // ゲームジャンル関係の変更
		v11(), // ゲームの公開範囲(visibility)の設定
		v12(), // ゲームジャンルとゲームの関係を後方参照に変更
		v13(), // gamesテーブルにバージョンの最終更新日時(latest_version_updated_at)カラムを追加
		v14(), // game_video_typesにmkvとm4vを追加
		v15(), // v2_game_versionsのnameカラムにunique制約を追加
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
