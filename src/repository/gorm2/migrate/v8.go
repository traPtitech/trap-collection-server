package migrate

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// v8
// v2 api用のデータのマイグレーション
func v8() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "8",
		Migrate: func(tx *gorm.DB) error {
			// ゲームファイル
			err := tx.Exec("INSERT INTO v2_game_files " +
				"(id, game_id, file_type_id, hash, entry_point, created_at) " +
				"SELECT game_files.id, game_versions.game_id, game_files.file_type_id, game_files.hash, game_files.entry_point, game_files.created_at " +
				"FROM game_files " +
				"LEFT JOIN game_versions ON game_files.game_version_id = game_versions.id").Error
			if err != nil {
				return fmt.Errorf("failed to migrate v2_game_files table: %w", err)
			}

			// ゲーム画像
			err = tx.Exec("INSERT INTO v2_game_images " +
				"(id, game_id, image_type_id, created_at) " +
				"SELECT id, game_id, image_type_id, created_at " +
				"FROM game_images").Error
			if err != nil {
				return fmt.Errorf("failed to migrate v2_game_images table: %w", err)
			}

			// ゲーム動画
			err = tx.Exec("INSERT INTO v2_game_videos " +
				"(id, game_id, video_type_id, created_at) " +
				"SELECT id, game_id, video_type_id, created_at " +
				"FROM game_videos").Error
			if err != nil {
				return fmt.Errorf("failed to migrate v2_game_videos table: %w", err)
			}

			// ゲームバージョン
			err = tx.Exec("INSERT INTO v2_game_versions " +
				"(id, game_id, game_image_id, game_video_id, name, description, url, created_at) " +
				"SELECT game_versions.id, game_versions.game_id, game_images.id, game_videos.id, game_versions.name, game_versions.description, game_urls.url, game_versions.created_at " +
				"FROM game_versions " +
				"LEFT JOIN games ON game_versions.game_id = games.id " +
				"LEFT JOIN (SELECT * FROM game_images AS gi1 WHERE id = (" +
				"SELECT id FROM game_images AS gi2 WHERE gi1.game_id = gi2.game_id ORDER BY gi2.created_at LIMIT 1" +
				")) AS game_images ON games.id = game_images.game_id " +
				"LEFT JOIN (SELECT * FROM game_videos AS gv1 WHERE id = (" +
				"SELECT id FROM game_videos AS gv2 WHERE gv1.game_id = gv2.game_id ORDER BY gv2.created_at LIMIT 1)" +
				") AS game_videos ON games.id = game_videos.game_id " +
				"LEFT JOIN game_urls ON game_versions.id = game_urls.game_version_id").Error
			if err != nil {
				return fmt.Errorf("failed to migrate v2_game_versions table: %w", err)
			}

			// ゲームバージョンとファイルの紐付け
			err = tx.Exec("INSERT INTO game_version_game_file_relations " +
				"(game_version_id, game_file_id) " +
				"SELECT game_version_id, id " +
				"FROM game_files").Error
			if err != nil {
				return fmt.Errorf("failed to migrate game_version_game_file_relations table: %w", err)
			}

			// エディション
			err = tx.Exec("INSERT INTO editions " +
				"(id, name, questionnaire_url, created_at, deleted_at) " +
				"SELECT id, name, questionnaire_url, created_at, deleted_at " +
				"FROM launcher_versions").Error
			if err != nil {
				return fmt.Errorf("failed to migrate editions table: %w", err)
			}

			// エディションとゲームバージョンの紐付け
			err = tx.Exec("INSERT INTO edition_game_version_relations " +
				"(edition_id, game_version_id) " +
				"SELECT launcher_version_game_relations.launcher_version_table_id, game_versions.id " +
				"FROM launcher_version_game_relations " +
				"LEFT JOIN (SELECT * FROM game_versions AS gv1 WHERE id = (SELECT id FROM game_versions AS gv2 WHERE gv1.game_id = gv2.game_id ORDER BY gv2.created_at LIMIT 1)) AS game_versions ON launcher_version_game_relations.game_table_id = game_versions.game_id").Error
			if err != nil {
				return fmt.Errorf("failed to migrate edition_game_version_relations table: %w", err)
			}

			// プロダクトキー
			err = tx.Exec("INSERT INTO product_keys "+
				"(id, edition_id, status_id, product_key, created_at) "+
				"SELECT id, launcher_version_id, IF(deleted_at IS NULL, "+
				"(SELECT id FROM product_key_statuses WHERE name = ?), "+
				"(SELECT id FROM product_key_statuses WHERE name = ?)"+
				"), product_key, created_at "+
				"FROM launcher_users", productKeyStatusActiveV6, productKeyStatusInactiveV6).Error
			if err != nil {
				return fmt.Errorf("failed to migrate product_keys table: %w", err)
			}

			// アクセストークン
			err = tx.Exec("INSERT INTO access_tokens " +
				"(id, product_key_id, access_token, expires_at, created_at, deleted_at) " +
				"SELECT id, launcher_user_id, access_token, expires_at, created_at, deleted_at " +
				"FROM launcher_sessions").Error
			if err != nil {
				return fmt.Errorf("failed to migrate access_tokens table: %w", err)
			}

			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	}
}
