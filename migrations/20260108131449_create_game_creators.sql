-- Create "game_creator_custom_jobs" table
CREATE TABLE `game_creator_custom_jobs` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `display_name` varchar(64) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `idx_game_creator_custom_jobs_game_id` (`game_id`),
  CONSTRAINT `fk_game_creator_custom_jobs_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_creators" table
CREATE TABLE `game_creators` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `user_name` varchar(32) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `idx_game_creators_game_id` (`game_id`),
  UNIQUE INDEX `idx_unique_game_id_user_id` (`game_id`, `user_id`),
  CONSTRAINT `fk_game_creators_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_creator_custom_job_relations" table
CREATE TABLE `game_creator_custom_job_relations` (
  `game_creator_id` varchar(36) NOT NULL,
  `custom_job_id` varchar(36) NOT NULL,
  PRIMARY KEY (`game_creator_id`, `custom_job_id`),
  INDEX `fk_game_creator_custom_job_relations_game_creator_customa6516cf7` (`custom_job_id`),
  CONSTRAINT `fk_game_creator_custom_job_relations_game_creator_customa6516cf7` FOREIGN KEY (`custom_job_id`) REFERENCES `game_creator_custom_jobs` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_creator_custom_job_relations_game_creator_table` FOREIGN KEY (`game_creator_id`) REFERENCES `game_creators` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_creator_jobs" table
CREATE TABLE `game_creator_jobs` (
  `id` varchar(36) NOT NULL,
  `display_name` varchar(64) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_creator_job_relations" table
CREATE TABLE `game_creator_job_relations` (
  `game_creator_id` varchar(36) NOT NULL,
  `job_id` varchar(36) NOT NULL,
  PRIMARY KEY (`game_creator_id`, `job_id`),
  INDEX `fk_game_creator_job_relations_game_creator_job_table` (`job_id`),
  CONSTRAINT `fk_game_creator_job_relations_game_creator_job_table` FOREIGN KEY (`job_id`) REFERENCES `game_creator_jobs` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_creator_job_relations_game_creator_table` FOREIGN KEY (`game_creator_id`) REFERENCES `game_creators` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
