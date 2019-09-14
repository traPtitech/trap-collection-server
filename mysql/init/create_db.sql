CREATE TABLE game (id varchar(36),name varchar(30),cantainer text, file_name text,created_at datetime,updated_at datetime);
CREATE TABLE administrators (user_traqid char(30));
CREATE TABLE options (id int(11),question_id int(11),option_num int(11),body text);
CREATE TABLE question (id int(11),questionnaire_id int(11),page_num int(11),question_num int(11),type char(20),body text,is_required tinyint(4),deleted_at timestamp,created_at timestamp);
CREATE TABLE questionnaires (id int(11),title char(50),description text,res_time_limit timestamp,deleted_at timestamp,res_shared_to char(30),created_at timestamp,modified_at timestamp);
CREATE TABLE respondents (response_id int(11),questionnaire_id int(11),modified_at timestamp,submitted_at timestamp,deleted_at timestamp);
CREATE TABLE response (response_id int(11),question_id int(11),body text,modified_at timestamp,deleted_at timestamp);
CREATE TABLE scale_labels (question_id int(11),scale_label_left text,scale_label_right text,scale_min int(11),scale_max int(11));