-- Modify "feedback_questions" table
ALTER TABLE `feedback_questions` MODIFY COLUMN `question_order` bigint NOT NULL, DROP COLUMN `deleted_at`;
-- Modify "game_feedback_answers" table
ALTER TABLE `game_feedback_answers` MODIFY COLUMN `answer` bigint NOT NULL;
-- Modify "v2_latest_game_version_times" table
ALTER TABLE `v2_latest_game_version_times` DROP FOREIGN KEY `fk_v2_latest_game_version_times_games`, ADD CONSTRAINT `fk_v2_latest_game_version_times_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON UPDATE CASCADE ON DELETE CASCADE;
