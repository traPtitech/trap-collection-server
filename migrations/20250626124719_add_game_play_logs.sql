-- Create "game_play_logs" table
CREATE TABLE `game_play_logs` (
  `id` varchar(36) NOT NULL,
  `edition_id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `game_version_id` varchar(36) NOT NULL,
  `start_time` datetime NOT NULL DEFAULT (current_timestamp()),
  `end_time` datetime NULL,
  `duration` int NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `updated_at` datetime NOT NULL DEFAULT (current_timestamp()) ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  INDEX `idx_game_play_logs_edition_id` (`edition_id`),
  INDEX `idx_game_play_logs_game_id` (`game_id`),
  INDEX `idx_game_play_logs_game_version_id` (`game_version_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
