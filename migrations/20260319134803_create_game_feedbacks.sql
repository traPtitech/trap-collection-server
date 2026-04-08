-- Create "feedback_questions" table
CREATE TABLE `feedback_questions` (
  `id` varchar(36) NOT NULL,
  `game_id` varchar(36) NOT NULL,
  `question_text` varchar(256) NOT NULL,
  `answer_type` tinyint NOT NULL,
  `question_order` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  `archived_at` datetime NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_feedback_questions_game_id` (`game_id`),
  CONSTRAINT `fk_feedback_questions_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_feedbacks" table
CREATE TABLE `game_feedbacks` (
  `id` varchar(36) NOT NULL,
  `game_version_id` varchar(36) NOT NULL,
  `comment` text NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `idx_game_feedbacks_game_version_id` (`game_version_id`),
  CONSTRAINT `fk_game_feedbacks_game_version` FOREIGN KEY (`game_version_id`) REFERENCES `v2_game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_feedback_answers" table
CREATE TABLE `game_feedback_answers` (
  `id` varchar(36) NOT NULL,
  `feedback_id` varchar(36) NOT NULL,
  `question_id` varchar(36) NOT NULL,
  `answer` bigint NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_game_feedback_answers_feedback_id` (`feedback_id`),
  UNIQUE INDEX `idx_game_feedback_answers_feedback_question` (`feedback_id`, `question_id`),
  INDEX `idx_game_feedback_answers_question_id` (`question_id`),
  CONSTRAINT `fk_game_feedback_answers_question` FOREIGN KEY (`question_id`) REFERENCES `feedback_questions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_feedbacks_answers` FOREIGN KEY (`feedback_id`) REFERENCES `game_feedbacks` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_feedback_configs" table
CREATE TABLE `game_feedback_configs` (
  `game_id` varchar(36) NOT NULL,
  `enabled` bool NOT NULL DEFAULT 0,
  PRIMARY KEY (`game_id`),
  CONSTRAINT `fk_game_feedback_configs_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Modify "v2_latest_game_version_times" table
ALTER TABLE `v2_latest_game_version_times` DROP FOREIGN KEY `fk_v2_latest_game_version_times_games`, ADD CONSTRAINT `fk_v2_latest_game_version_times_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE CASCADE ON DELETE CASCADE;
