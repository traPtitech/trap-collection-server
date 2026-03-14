-- Create "v2_latest_game_version_times" table
CREATE TABLE `v2_latest_game_version_times` (
  `game_id` varchar(36) NOT NULL,
  `latest_game_version_id` varchar(36) NOT NULL,
  `latest_game_version_created_at` datetime NOT NULL,
  PRIMARY KEY (`game_id`),
  INDEX `idx_game_version_stats_latest_created_at` (`latest_game_version_created_at`),
  CONSTRAINT `fk_v2_latest_game_version_times_games` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE CASCADE ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- ここから手書き
-- 最新ゲームバージョン更新時間データの移行
INSERT INTO v2_latest_game_version_times (game_id, latest_game_version_id, latest_game_version_created_at)
SELECT 
    game_id,
    id,
    created_at
    FROM (
        SELECT 
            game_id,
            id,
            created_at,
            ROW_NUMBER() OVER (PARTITION BY game_id ORDER BY created_at DESC, id DESC) AS num
        FROM v2_game_versions
    ) AS ranked_versions
WHERE num = 1;
-- 元データ（latest_version_updated_atカラム）の無効化
ALTER TABLE `games` MODIFY COLUMN `latest_version_updated_at` datetime NULL;
