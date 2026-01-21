-- Create "feedback_questions" table
CREATE TABLE `feedback_questions` (
  `id` varchar(36) NOT NULL,
  `question_text` varchar(256) NOT NULL,
  `question_order` bigint NOT NULL,
  `is_active` bool NOT NULL DEFAULT 1,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_feedbacks" table
CREATE TABLE `game_feedbacks` (
  `id` varchar(36) NOT NULL,
  `edition_id` varchar(36) NOT NULL,
  `game_version_id` varchar(36) NOT NULL,
  `comment` text NULL,
  `created_at` datetime NOT NULL DEFAULT (current_timestamp()),
  PRIMARY KEY (`id`),
  INDEX `idx_game_feedbacks_edition_id` (`edition_id`),
  INDEX `idx_game_feedbacks_game_version_id` (`game_version_id`),
  CONSTRAINT `fk_game_feedbacks_edition` FOREIGN KEY (`edition_id`) REFERENCES `editions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_feedbacks_game_version` FOREIGN KEY (`game_version_id`) REFERENCES `v2_game_versions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "game_feedback_answers" table
CREATE TABLE `game_feedback_answers` (
  `id` varchar(36) NOT NULL,
  `feedback_id` varchar(36) NOT NULL,
  `question_id` varchar(36) NOT NULL,
  `answer` bool NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_game_feedback_answers_feedback_id` (`feedback_id`),
  INDEX `idx_game_feedback_answers_question_id` (`question_id`),
  CONSTRAINT `fk_game_feedback_answers_question` FOREIGN KEY (`question_id`) REFERENCES `feedback_questions` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT,
  CONSTRAINT `fk_game_feedbacks_answers` FOREIGN KEY (`feedback_id`) REFERENCES `game_feedbacks` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
