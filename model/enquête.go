package model

import (
	"time"

	"github.com/go-sql-driver/mysql"
)

//Questionnaires アンケートセットの構造体
type Questionnaires struct {
	ID           string         `json:"questionnaireID" db:"id"`
	Title        string         `json:"title"           db:"title"`
	Description  string         `json:"description"     db:"description"`
	ResTimeLimit mysql.NullTime `json:"res_time_limit"  db:"res_time_limit"`
	DeletedAt    mysql.NullTime `json:"deleted_at"      db:"deleted_at"`
	CreatedAt    time.Time      `json:"created_at"      db:"created_at"`
	ModifiedAt   time.Time      `json:"modified_at"     db:"modified_at"`
}

//QuestionnairesInfo アンケートセットの詳細の構造体
type QuestionnairesInfo struct {
	ID           string `json:"questionnaireID"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ResTimeLimit string `json:"res_time_limit"`
	CreatedAt    string `json:"created_at"`
	ModifiedAt   string `json:"modified_at"`
}

//Questions 質問の構造体
type Questions struct {
	ID              string         `json:"id"                  db:"id"`
	QuestionnaireID string         `json:"questionnaireID"     db:"questionnaire_id"`
	PageNum         int            `json:"page_num"            db:"page_num"`
	QuestionNum     int            `json:"question_num"        db:"question_num"`
	Type            string         `json:"type"                db:"type"`
	Body            string         `json:"body"                db:"body"`
	IsRequrired     bool           `json:"is_required"         db:"is_required"`
	DeletedAt       mysql.NullTime `json:"deleted_at"          db:"deleted_at"`
	CreatedAt       time.Time      `json:"created_at"          db:"created_at"`
}

//QuestionIDType 質問のIDと種類の構造体
type QuestionIDType struct {
	ID   string `db:"id"`
	Type string `db:"type"`
}

//ScaleLabels メモリ形式の質問の左右の値
type ScaleLabels struct {
	ID              string `json:"questionID" db:"question_id"`
	ScaleLabelRight string `json:"scale_label_right" db:"scale_label_right"`
	ScaleLabelLeft  string `json:"scale_label_left"  db:"scale_label_left"`
	ScaleMin        int    `json:"scale_min" db:"scale_min"`
	ScaleMax        int    `json:"scale_max" db:"scale_max"`
}

//ResponseBody 回答の構造体
type ResponseBody struct {
	QuestionID     string   `json:"questionID"`
	QuestionType   string   `json:"question_type"`
	Response       string   `json:"response"`
	OptionResponse []string `json:"option_response"`
}

//Responses 回答の構造体
type Responses struct {
	ID          string         `json:"questionnaireID"`
	SubmittedAt string         `json:"submitted_at"`
	Body        []ResponseBody `json:"body"`
}

//ResponseInfo 回答の構造体
type ResponseInfo struct {
	QuestionnaireID string         `db:"questionnaire_id"`
	ResponseID      string         `db:"response_id"`
	ModifiedAt      time.Time      `db:"modified_at"`
	SubmittedAt     mysql.NullTime `db:"submitted_at"`
}

//MyResponse 回答の構造体
type MyResponse struct {
	ResponseID      string `json:"responseID"`
	QuestionnaireID string `json:"questionnaireID"`
	Title           string `json:"questionnaire_title"`
	ResTimeLimit    string `json:"res_time_limit"`
	SubmittedAt     string `json:"submitted_at"`
	ModifiedAt      string `json:"modified_at"`
}

//ResponseID 回答の構造体
type ResponseID struct {
	QuestionnaireID string         `db:"questionnaire_id"`
	ModifiedAt      mysql.NullTime `db:"modified_at"`
	SubmittedAt     mysql.NullTime `db:"submitted_at"`
}

//QIDandResponse アンケートのIDと回答の構造体
type QIDandResponse struct {
	QuestionID string
	Response   string
}

//ResponseAnDBody 回答と内容の構造体
type ResponseAnDBody struct {
	ResponseID  string         `db:"response_id"`
	ModifiedAt  time.Time      `db:"modified_at"`
	SubmittedAt mysql.NullTime `db:"submitted_at"`
	QuestionID  string         `db:"question_id"`
	Body        string         `db:"body"`
}

//UserResponse 回答の構造体
type UserResponse struct {
	ResponseID  string         `db:"response_id"`
	ModifiedAt  time.Time      `db:"modified_at"`
	SubmittedAt mysql.NullTime `db:"submitted_at"`
}
