DROP DATABASE IF EXISTS trap_collection;
CREATE DATABASE trap_collection;
USE trap_collection;

CREATE TABLE game (id varchar(36),name varchar(30),container text, file_name text,md5 binary(16),time timestamp default NULL,created_at timestamp default NULL,updated_at timestamp default NULL,deleted_at timestamp default NULL);
CREATE TABLE versions_for_sale (id varchar(36),name varchar(30),start_period timestamp default NULL,end_period timestamp default NULL,start_time timestamp default NULL,created_at timestamp default NULL,updated_at timestamp default NULL,deleted_at timestamp default NULL);
CREATE TABLE versions_not_for_sale (id varchar(36),name varchar(30),questionnaire_id varchar(36),start_period timestamp default NULL,end_period timestamp default NULL,start_time timestamp default NULL,created_at timestamp default NULL,updated_at timestamp default NULL,deleted_at timestamp default NULL);
CREATE TABLE seat (id varchar(36),seat_id varchar(36),created_at timestamp default NULL,deleted_at timestamp default NULL);
CREATE TABLE play_time (id varchar(36),version_id varchar(36),game_id varchar(36),start_time timestamp default NULL,end_time timestamp default NULL);
CREATE TABLE special (id varchar(36),version_id varchar(30),game_name varchar(30),status text,deleted_at timestamp default NULL);
CREATE TABLE administrators (user_traqid char(30));
CREATE TABLE options (id varchar(36),question_id varchar(36),option_num varchar(36),body text);
CREATE TABLE question (id varchar(36),questionnaire_id varchar(36),page_num varchar(36),question_num varchar(36),type char(20),body text,is_required tinyint(4),deleted_at timestamp default NULL,created_at timestamp default NULL);
CREATE TABLE questionnaires (id varchar(36),title char(50),description text,res_time_limit timestamp default NULL,deleted_at timestamp default NULL,created_at timestamp default NULL,modified_at timestamp default NULL);
CREATE TABLE respondents (response_id varchar(36),questionnaire_id varchar(36),modified_at timestamp default NULL,submitted_at timestamp default NULL,deleted_at timestamp default NULL);
CREATE TABLE response (response_id varchar(36),question_id varchar(36),body text,modified_at timestamp default NULL,deleted_at timestamp default NULL);
CREATE TABLE scale_labels (question_id varchar(36),scale_label_left text,scale_label_right text,scale_min varchar(36),scale_max varchar(36));