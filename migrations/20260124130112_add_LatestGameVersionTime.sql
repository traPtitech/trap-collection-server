-- Create "v2_latest_game_version_times" table
CREATE TABLE `v2_latest_game_version_times` (
  `game_id` varchar(36) NOT NULL,
  `latest_game_version_id` varchar(36) NOT NULL,
  `latest_game_version_created_at` datetime NOT NULL,
  PRIMARY KEY (`game_id`),
  INDEX `idx_game_version_stats_latest_created_at` (`latest_game_version_created_at`)
  CONSTRAINT `fk_v2_latest_game_version_times_games` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE CASCADE ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- ここから手書き
INSERT INTO v2_latest_game_version_times (game_id, latest_game_version_id, latest_game_version_created_at)
SELECT 
    v1.game_id,
    v1.id,
    v1.created_at
FROM 
    v2_game_versions AS v1
JOIN (
    SELECT game_id, MAX(created_at) AS max_created_at
    FROM v2_game_versions
    GROUP BY game_id
) AS v2 ON v1.game_id = v2.game_id AND v1.created_at = v2.max_created_at
ON DUPLICATE KEY UPDATE
    latest_game_version_id = VALUES(latest_game_version_id),
    latest_game_version_created_at = VALUES(latest_game_version_created_at);
ALTER TABLE `games` MODIFY COLUMN `latest_version_updated_at` datetime NULL;
