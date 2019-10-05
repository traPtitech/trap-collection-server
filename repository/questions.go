package repository

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
)

//GetQuestionsType 質問のIDと種類の取得
func GetQuestionsType(c echo.Context, questionnaireID string) ([]model.QuestionIDType, error) {
	ret := []model.QuestionIDType{}
	if err := Db.Select(&ret,
		"SELECT id, type FROM question WHERE questionnaire_id = ? AND deleted_at IS NULL ORDER BY question_num",
		questionnaireID); err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return ret, nil
}

//GetQuestions 質問の取得
func GetQuestions(c echo.Context, questionnaireID string) ([]model.Questions, error) {
	allquestions := []model.Questions{}

	// アンケートidの一致する質問を取る
	if err := Db.Select(&allquestions,
		"SELECT * FROM question WHERE questionnaire_id = ? AND deleted_at IS NULL ORDER BY question_num",
		questionnaireID); err != nil {
		c.Logger().Error(err)
		return []model.Questions{}, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return allquestions, nil
}

//InsertQuestion 質問の追加
func InsertQuestion(
	c echo.Context, questionnaireID string, pageNum int, questionNum int, questionType string,
	body string, isRequired bool) (string, error) {
	lastID := uuid.Must(uuid.NewV4()).String()
	_, err := Db.Exec(
		`INSERT INTO question (id,questionnaire_id, page_num, question_num, type, body, is_required, created_at)
		VALUES (?,?, ?, ?, ?, ?, ?, ?)`,
		lastID, questionnaireID, pageNum, questionNum, questionType, body, isRequired, time.Now())
	if err != nil {
		c.Logger().Error(err)
		return "", echo.NewHTTPError(http.StatusInternalServerError)
	}

	return lastID, nil
}

//UpdateQuestion 質問の変更
func UpdateQuestion(
	c echo.Context, questionnaireID string, pageNum int, questionNum int, questionType string,
	body string, isRequired bool, questionID string) error {
	if _, err := Db.Exec(
		"UPDATE question SET questionnaire_id = ?, page_num = ?, question_num = ?, type = ?, body = ?, is_required = ? WHERE id = ?",
		questionnaireID, pageNum, questionNum, questionType, body, isRequired, questionID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//DeleteQuestion 質問の削除
func DeleteQuestion(c echo.Context, questionID string) error {
	if _, err := Db.Exec(
		"UPDATE question SET deleted_at = ? WHERE id = ?",
		time.Now(), questionID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//GetResShared アンケートの公開範囲の取得
func GetResShared(c echo.Context, questionnaireID string) (string, error) {
	resSharedTo := ""
	if err := Db.Get(&resSharedTo,
		"SELECT res_shared_to FROM questionnaires WHERE deleted_at IS NULL AND id = ?",
		questionnaireID); err != nil {
		c.Logger().Error(err)
		if err == sql.ErrNoRows {
			return "", echo.NewHTTPError(http.StatusNotFound)
		}
		return "", echo.NewHTTPError(http.StatusInternalServerError)
	}
	return resSharedTo, nil
}
