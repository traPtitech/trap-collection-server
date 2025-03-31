-- Create "admins" table
CREATE TABLE `admins` (
  `user_id` varchar(36) NOT NULL,
  PRIMARY KEY (`user_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "editions" table
CREATE TABLE `editions` (
  `id` varchar(36) NOT NULL,
  `name` varchar(32) NOT NULL,
  `questionnaire_url` text NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `deleted_at` datetime NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_editions_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "migrations" table
CREATE TABLE `migrations` (
  `id` varchar(190) NOT NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "product_key_statuses" table
CREATE TABLE `product_key_statuses` (
  `id` tinyint NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `active` bool NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_product_key_statuses_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "product_keys" table
CREATE TABLE `product_keys` (
  `id` varchar(36) NOT NULL,
  `edition_id` varchar(36) NOT NULL,
  `product_key` varchar(29) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `status_id` tinyint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_editions_product_keys` (`edition_id`),
  INDEX `fk_product_keys_status` (`status_id`),
  UNIQUE INDEX `uni_product_keys_product_key` (`product_key`),
  CONSTRAINT `fk_editions_product_keys` FOREIGN KEY (`edition_id`) REFERENCES `editions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_product_keys_status` FOREIGN KEY (`status_id`) REFERENCES `product_key_statuses` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "access_tokens" table
CREATE TABLE `access_tokens` (
  `id` varchar(36) NOT NULL,
  `product_key_id` varchar(36) NOT NULL,
  `access_token` varchar(64) NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `deleted_at` datetime NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_product_keys_access_tokens` (`product_key_id`),
  UNIQUE INDEX `uni_access_tokens_access_token` (`access_token`),
  CONSTRAINT `fk_product_keys_access_tokens` FOREIGN KEY (`product_key_id`) REFERENCES `product_keys` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_visibility_types" table
CREATE TABLE `game_visibility_types` (
  `id` tinyint NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_game_visibility_types_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "games" table
CREATE TABLE `games` (
  `id` varchar(36) NOT NULL,
  `name` varchar(256) NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `deleted_at` datetime NULL,
  `visibility_type_id` tinyint NOT NULL,
  `latest_version_updated_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_games_game_visibility_type` (`visibility_type_id`),
  CONSTRAINT `fk_games_game_visibility_type` FOREIGN KEY (`visibility_type_id`) REFERENCES `game_visibility_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_image_types" table
CREATE TABLE `game_image_types` (
  `id` tinyint NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `active` bool NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_game_image_types_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "v2_game_images" table
CREATE TABLE `v2_game_images` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `image_type_id` tinyint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_games_game_image2s` (`game_id`),
  INDEX `fk_v2_game_images_game_image_type` (`image_type_id`),
  CONSTRAINT `fk_games_game_image2s` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_v2_game_images_game_image_type` FOREIGN KEY (`image_type_id`) REFERENCES `game_image_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_video_types" table
CREATE TABLE `game_video_types` (
  `id` tinyint NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `active` bool NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_game_video_types_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "v2_game_videos" table
CREATE TABLE `v2_game_videos` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `video_type_id` tinyint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_games_game_video2s` (`game_id`),
  INDEX `fk_v2_game_videos_game_video_type` (`video_type_id`),
  CONSTRAINT `fk_games_game_video2s` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_v2_game_videos_game_video_type` FOREIGN KEY (`video_type_id`) REFERENCES `game_video_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "v2_game_versions" table
CREATE TABLE `v2_game_versions` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `game_image_id` varchar(36) NOT NULL,
  `game_video_id` varchar(36) NOT NULL,
  `name` varchar(32) NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `url` text NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_v2_game_versions_game_image` (`game_image_id`),
  INDEX `fk_v2_game_versions_game_video` (`game_video_id`),
  UNIQUE INDEX `idx_game_id_name` (`game_id`, `name`),
  CONSTRAINT `fk_v2_game_versions_game_image` FOREIGN KEY (`game_image_id`) REFERENCES `v2_game_images` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_v2_game_versions_game_video` FOREIGN KEY (`game_video_id`) REFERENCES `v2_game_videos` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "edition_game_version_relations" table
CREATE TABLE `edition_game_version_relations` (
  `edition_id` varchar(36) NOT NULL,
  `game_version_id` varchar(36) NOT NULL,
  PRIMARY KEY (`edition_id`, `game_version_id`),
  INDEX `fk_edition_game_version_relations_game_version_table2_v2` (`game_version_id`),
  CONSTRAINT `fk_edition_game_version_relations_edition_table_v2` FOREIGN KEY (`edition_id`) REFERENCES `editions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_edition_game_version_relations_game_version_table2_v2` FOREIGN KEY (`game_version_id`) REFERENCES `v2_game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_file_types" table
CREATE TABLE `game_file_types` (
  `id` tinyint NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `active` bool NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_game_file_types_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_versions" table
CREATE TABLE `game_versions` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `name` varchar(32) NOT NULL,
  `description` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_games_game_versions` (`game_id`),
  CONSTRAINT `fk_games_game_versions` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_files" table
CREATE TABLE `game_files` (
  `id` varchar(36) NOT NULL,
  `game_version_id` varchar(36) NOT NULL,
  `file_type_id` tinyint NOT NULL,
  `hash` char(32) NOT NULL,
  `entry_point` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_game_files_game_file_type` (`file_type_id`),
  UNIQUE INDEX `idx_game_file_unique` (`game_version_id`, `file_type_id`),
  CONSTRAINT `fk_game_files_game_file_type` FOREIGN KEY (`file_type_id`) REFERENCES `game_file_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_versions_game_files` FOREIGN KEY (`game_version_id`) REFERENCES `game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_genres" table
CREATE TABLE `game_genres` (
  `id` varchar(36) NOT NULL,
  `name` varchar(32) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_game_genres_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_genre_relations" table
CREATE TABLE `game_genre_relations` (
  `genre_id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  PRIMARY KEY (`genre_id`, `game_id`),
  INDEX `fk_game_genre_relations_game_table2_v13` (`game_id`),
  CONSTRAINT `fk_game_genre_relations_game_genre_table_v10` FOREIGN KEY (`genre_id`) REFERENCES `game_genres` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_genre_relations_game_genre_table_v11` FOREIGN KEY (`genre_id`) REFERENCES `game_genres` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_genre_relations_game_genre_table_v12` FOREIGN KEY (`genre_id`) REFERENCES `game_genres` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_genre_relations_game_genre_table_v13` FOREIGN KEY (`genre_id`) REFERENCES `game_genres` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_genre_relations_game_table2_v11` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_genre_relations_game_table2_v12` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_genre_relations_game_table2_v13` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_genre_relations_game_table2_v5` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_images" table
CREATE TABLE `game_images` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `image_type_id` tinyint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_games_game_images` (`game_id`),
  INDEX `fk_game_images_game_image_type` (`image_type_id`),
  CONSTRAINT `fk_game_images_game_image_type` FOREIGN KEY (`image_type_id`) REFERENCES `game_image_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_games_game_images` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_management_role_types" table
CREATE TABLE `game_management_role_types` (
  `id` tinyint NOT NULL AUTO_INCREMENT,
  `name` varchar(32) NOT NULL,
  `active` bool NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_game_management_role_types_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_management_roles" table
CREATE TABLE `game_management_roles` (
  `game_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `role_type_id` tinyint NOT NULL,
  PRIMARY KEY (`game_id`, `user_id`),
  INDEX `fk_game_management_roles_role_type_table` (`role_type_id`),
  CONSTRAINT `fk_game_management_roles_role_type_table` FOREIGN KEY (`role_type_id`) REFERENCES `game_management_role_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_games_game_management_roles` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_urls" table
CREATE TABLE `game_urls` (
  `id` varchar(36) NOT NULL,
  `game_version_id` varchar(36) NOT NULL,
  `url` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_game_urls_game_version_id` (`game_version_id`),
  CONSTRAINT `fk_game_versions_game_url` FOREIGN KEY (`game_version_id`) REFERENCES `game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "v2_game_files" table
CREATE TABLE `v2_game_files` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `file_type_id` tinyint NOT NULL,
  `hash` char(32) NOT NULL,
  `entry_point` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_games_game_files` (`game_id`),
  INDEX `fk_v2_game_files_game_file_type` (`file_type_id`),
  CONSTRAINT `fk_games_game_files` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_v2_game_files_game_file_type` FOREIGN KEY (`file_type_id`) REFERENCES `game_file_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_version_game_file_relations" table
CREATE TABLE `game_version_game_file_relations` (
  `game_version_id` varchar(36) NOT NULL,
  `game_file_id` varchar(36) NOT NULL,
  PRIMARY KEY (`game_version_id`, `game_file_id`),
  INDEX `fk_game_version_game_file_relations_game_file_table2_v5` (`game_file_id`),
  CONSTRAINT `fk_game_version_game_file_relations_game_file_table2_v2` FOREIGN KEY (`game_file_id`) REFERENCES `v2_game_files` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_version_game_file_relations_game_file_table2_v5` FOREIGN KEY (`game_file_id`) REFERENCES `v2_game_files` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_version_game_file_relations_game_version_table2_v15` FOREIGN KEY (`game_version_id`) REFERENCES `v2_game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_version_game_file_relations_game_version_table2_v2` FOREIGN KEY (`game_version_id`) REFERENCES `v2_game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_version_game_file_relations_game_version_table2_v4` FOREIGN KEY (`game_version_id`) REFERENCES `v2_game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_videos" table
CREATE TABLE `game_videos` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `video_type_id` tinyint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `fk_games_game_videos` (`game_id`),
  INDEX `fk_game_videos_game_video_type` (`video_type_id`),
  CONSTRAINT `fk_game_videos_game_video_type` FOREIGN KEY (`video_type_id`) REFERENCES `game_video_types` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_games_game_videos` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "launcher_versions" table
CREATE TABLE `launcher_versions` (
  `id` varchar(36) NOT NULL,
  `name` varchar(32) NOT NULL,
  `questionnaire_url` text NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `deleted_at` datetime NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uni_launcher_versions_name` (`name`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "launcher_users" table
CREATE TABLE `launcher_users` (
  `id` varchar(36) NOT NULL,
  `launcher_version_id` varchar(36) NOT NULL,
  `product_key` varchar(29) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `deleted_at` datetime NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_launcher_versions_launcher_users` (`launcher_version_id`),
  UNIQUE INDEX `uni_launcher_users_product_key` (`product_key`),
  CONSTRAINT `fk_launcher_versions_launcher_users` FOREIGN KEY (`launcher_version_id`) REFERENCES `launcher_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "launcher_sessions" table
CREATE TABLE `launcher_sessions` (
  `id` varchar(36) NOT NULL,
  `launcher_user_id` varchar(36) NOT NULL,
  `access_token` varchar(64) NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `deleted_at` datetime NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_launcher_users_launcher_sessions` (`launcher_user_id`),
  UNIQUE INDEX `uni_launcher_sessions_access_token` (`access_token`),
  CONSTRAINT `fk_launcher_users_launcher_sessions` FOREIGN KEY (`launcher_user_id`) REFERENCES `launcher_users` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "launcher_version_game_relations" table
CREATE TABLE `launcher_version_game_relations` (
  `launcher_version_table_id` varchar(36) NOT NULL,
  `game_table_id` varchar(36) NOT NULL,
  PRIMARY KEY (`launcher_version_table_id`, `game_table_id`),
  INDEX `fk_launcher_version_game_relations_game_table` (`game_table_id`),
  CONSTRAINT `fk_launcher_version_game_relations_game_table` FOREIGN KEY (`game_table_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_launcher_version_game_relations_launcher_version_table` FOREIGN KEY (`launcher_version_table_id`) REFERENCES `launcher_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "seat_statuses" table
CREATE TABLE `seat_statuses` (
  `id` tinyint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `active` bool NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "seats" table
CREATE TABLE `seats` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `status_id` tinyint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `fk_seats_seat_status` (`status_id`),
  CONSTRAINT `fk_seats_seat_status` FOREIGN KEY (`status_id`) REFERENCES `seat_statuses` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
