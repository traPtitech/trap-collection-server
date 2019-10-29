DROP DATABASE IF EXISTS trap_collection;
CREATE DATABASE trap_collection;
USE trap_collection;

CREATE TABLE game (id varchar(36),name varchar(30),container text, file_name text,md5 binary(16),time timestamp NULL,created_at timestamp NULL,updated_at timestamp NULL,deleted_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE versions_for_sale (id varchar(36),name varchar(30),start_period timestamp NULL,end_period timestamp NULL,start_time timestamp NULL,created_at timestamp NULL,updated_at timestamp NULL,deleted_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE versions_not_for_sale (id varchar(36),name varchar(30),questionnaire_id varchar(36),start_period timestamp NULL,end_period timestamp NULL,start_time timestamp NULL,created_at timestamp NULL,updated_at timestamp NULL,deleted_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE seat (id varchar(36),seat_id varchar(36),created_at timestamp NULL,deleted_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE play_time (id varchar(36),version_id varchar(36),game_id varchar(36),start_time timestamp NULL,end_time timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE special (id varchar(36),version_id varchar(36),game_name varchar(30),status text,deleted_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE administrators (user_traqid char(30)) CHARACTER SET utf8mb4;
CREATE TABLE options (id varchar(36),question_id varchar(36),option_num int(11),body text) CHARACTER SET utf8mb4;
CREATE TABLE question (id varchar(36),questionnaire_id varchar(36),page_num varchar(36),question_num varchar(36),type char(20),body text,is_required tinyint(4),deleted_at timestamp NULL,created_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE questionnaires (id varchar(36),title char(50),description text,res_time_limit timestamp NULL,deleted_at timestamp NULL,created_at timestamp NULL,modified_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE respondents (response_id varchar(36),questionnaire_id varchar(36),modified_at timestamp NULL,submitted_at timestamp NULL,deleted_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE response (response_id varchar(36),question_id varchar(36),body text,modified_at timestamp NULL,deleted_at timestamp NULL) CHARACTER SET utf8mb4;
CREATE TABLE scale_labels (question_id varchar(36),scale_label_left text,scale_label_right text,scale_min varchar(36),scale_max varchar(36));