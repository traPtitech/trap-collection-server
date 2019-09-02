package repository

import (
	"net/http"
	"strconv"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/traPtitech/trap-collection-server/model"
)

// GetAllQuestionnaires エラーが起きれば(nil, err)
// 起こらなければ(allquestionnaires, nil)を返す
func GetAllQuestionnaires(c echo.Context) ([]model.Questionnaires, error) {
	// query parametar
	sort := c.QueryParam("sort")

	var list = map[string]string{
		"":             "",
		"created_at":   "ORDER BY created_at",
		"-created_at":  "ORDER BY created_at DESC",
		"title":        "ORDER BY title",
		"-title":       "ORDER BY title DESC",
		"modified_at":  "ORDER BY modified_at",
		"-modified_at": "ORDER BY modified_at DESC",
	}
	// アンケート一覧の配列
	allquestionnaires := []model.Questionnaires{}

	if err := Db.Select(&allquestionnaires,
		"SELECT * FROM questionnaires WHERE deleted_at IS NULL "+list[sort]); err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return allquestionnaires, nil
}

/*
GetQuestionnaires アンケートの一覧
2つ目の戻り値はページ数の最大
*/
func GetQuestionnaires(c echo.Context, nontargeted bool) ([]model.QuestionnairesInfo, int, error) {
	allquestionnaires, err := GetAllQuestionnaires(c)
	if err != nil {
		return nil, 0, err
	}

	questionnaires := []model.QuestionnairesInfo{}
	for _, q := range allquestionnaires {
		questionnaires = append(questionnaires,
			model.QuestionnairesInfo{
				ID:           q.ID,
				Title:        q.Title,
				Description:  q.Description,
				ResTimeLimit: NullTimeToString(q.ResTimeLimit),
				ResSharedTo:  q.ResSharedTo,
				CreatedAt:    q.CreatedAt.Format(time.RFC3339),
				ModifiedAt:   q.ModifiedAt.Format(time.RFC3339)})
	}

	if len(questionnaires) == 0 {
		return nil, 0, echo.NewHTTPError(http.StatusNotFound)
	}

	pageMax := len(questionnaires)/20 + 1

	page := c.QueryParam("page")
	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		c.Logger().Error(err)
		return nil, 0, echo.NewHTTPError(http.StatusBadRequest)
	}

	if pageNum > pageMax {
		return nil, 0, echo.NewHTTPError(http.StatusBadRequest)
	}

	ret := []model.QuestionnairesInfo{}
	for i := 0; i < 20; i++ {
		index := (pageNum-1)*20 + i
		if index >= len(questionnaires) {
			break
		}
		ret = append(ret, questionnaires[index])
	}

	return ret, pageMax, nil
}

//GetQuestionnaire アンケートの取得
func GetQuestionnaire(c echo.Context, questionnaireID int) (model.Questionnaires, error) {
	questionnaire := model.Questionnaires{}
	if err := Db.Get(&questionnaire, "SELECT * FROM questionnaires WHERE id = ? AND deleted_at IS NULL", questionnaireID); err != nil {
		c.Logger().Error(err)
		if err == sql.ErrNoRows {
			return model.Questionnaires{}, echo.NewHTTPError(http.StatusNotFound)
		}
		return model.Questionnaires{}, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return questionnaire, nil
}

//GetQuestionnaireInfo アンケートの詳細取得
func GetQuestionnaireInfo(c echo.Context, questionnaireID int) (model.Questionnaires, []string, error) {
	questionnaire, err := GetQuestionnaire(c, questionnaireID)
	if err != nil {
		return model.Questionnaires{}, nil, err
	}

	administrators, err := GetAdministrators(c, questionnaireID)
	if err != nil {
		return model.Questionnaires{}, nil, err
	}
	return questionnaire, administrators, nil
}

//GetQuestionnaireLimit アンケートの回答期限取得
func GetQuestionnaireLimit(c echo.Context, questionnaireID int) (string, error) {
	res := struct {
		Title        string         `Db:"title"`
		ResTimeLimit mysql.NullTime `Db:"res_time_limit"`
	}{}
	if err := Db.Get(&res,
		"SELECT res_time_limit FROM questionnaires WHERE id = ? AND deleted_at IS NULL",
		questionnaireID); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		c.Logger().Error(err)
		return "", echo.NewHTTPError(http.StatusInternalServerError)
	}
	return NullTimeToString(res.ResTimeLimit), nil
}

//GetTitleAndLimit アンケートのタイトルと回答期限取得
func GetTitleAndLimit(c echo.Context, questionnaireID int) (string, string, error) {
	res := struct {
		Title        string         `Db:"title"`
		ResTimeLimit mysql.NullTime `Db:"res_time_limit"`
	}{}
	if err := Db.Get(&res,
		"SELECT title, res_time_limit FROM questionnaires WHERE id = ? AND deleted_at IS NULL",
		questionnaireID); err != nil {
		if err == sql.ErrNoRows {
			return "", "", nil
		}
		c.Logger().Error(err)
		return "", "", echo.NewHTTPError(http.StatusInternalServerError)
	}
	return res.Title, NullTimeToString(res.ResTimeLimit), nil
}

//InsertQuestionnaire アンケートの追加
func InsertQuestionnaire(c echo.Context, title string, description string, resTimeLimit string, resSharedTo string) (int, error) {
	var result sql.Result

	if resTimeLimit == "" || resTimeLimit == "NULL" {
		resTimeLimit = "NULL"
		var err error
		result, err = Db.Exec(
			`INSERT INTO questionnaires (title, description, res_shared_to, created_at, modified_at)
			VALUES (?, ?, ?, ?, ?)`,
			title, description, resSharedTo, time.Now(), time.Now())
		if err != nil {
			c.Logger().Error(err)
			return 0, echo.NewHTTPError(http.StatusInternalServerError)
		}
	} else {
		var err error
		result, err = Db.Exec(
			`INSERT INTO questionnaires (title, description, res_time_limit, res_shared_to, created_at, modified_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			title, description, resTimeLimit, resSharedTo, time.Now(), time.Now())
		if err != nil {
			c.Logger().Error(err)
			return 0, echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		c.Logger().Error(err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError)
	}

	return int(lastID), nil
}

//UpdateQuestionnaire アンケートの追加
func UpdateQuestionnaire(c echo.Context, title string, description string, resTimeLimit string, resSharedTo string, questionnaireID int) error {
	if resTimeLimit == "" || resTimeLimit == "NULL" {
		resTimeLimit = "NULL"
		if _, err := Db.Exec(
			`UPDATE questionnaires SET title = ?, description = ?, res_time_limit = NULL,
			res_shared_to = ?, modified_at = ? WHERE id = ?`,
			title, description, resSharedTo, time.Now(), questionnaireID); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	} else {
		if _, err := Db.Exec(
			`UPDATE questionnaires SET title = ?, description = ?, res_time_limit = ?,
			res_shared_to = ?, modified_at = ? WHERE id = ?`,
			title, description, resTimeLimit, resSharedTo, time.Now(), questionnaireID); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return nil
}

//DeleteQuestionnaire アンケートの削除
func DeleteQuestionnaire(c echo.Context, questionnaireID int) error {
	if _, err := Db.Exec(
		"UPDATE questionnaires SET deleted_at = ? WHERE id = ?",
		time.Now(), questionnaireID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}
