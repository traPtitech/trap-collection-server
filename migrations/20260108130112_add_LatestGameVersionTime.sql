-- Create "v2_latest_game_version_times" table
CREATE TABLE `v2_latest_game_version_times` (
  `game_id` varchar(36) NOT NULL,
  `latest_game_version_id` varchar(36) NOT NULL,
  `latest_game_version_created_at` datetime NOT NULL,
  PRIMARY KEY (`game_id`),
  INDEX `idx_game_version_stats_latest_created_at` (`latest_game_version_created_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
