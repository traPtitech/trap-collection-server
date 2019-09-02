package repository

import (
	"net/http"

	"github.com/labstack/echo"
)

//GetAdministrators 管理者の取得
func GetAdministrators(c echo.Context, questionnaireID int) ([]string, error) {
	administrators := []string{}
	if err := Db.Select(&administrators, "SELECT user_traqid FROM administrators WHERE questionnaire_id = ?", questionnaireID); err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return administrators, nil
}

//InsertAdministrators 管理者の追加
func InsertAdministrators(c echo.Context, questionnaireID int, administrators []string) error {
	for _, v := range administrators {
		if _, err := Db.Exec(
			"INSERT INTO administrators (questionnaire_id, user_traqid) VALUES (?, ?)",
			questionnaireID, v); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return nil
}

//DeleteAdministrators 管理者の削除
func DeleteAdministrators(c echo.Context, questionnaireID int) error {
	if _, err := Db.Exec(
		"DELETE from administrators WHERE questionnaire_id = ?",
		questionnaireID); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//CheckAdmin 自分がadminなら(true, nil)
func CheckAdmin(c echo.Context, questionnaireID int) (bool, error) {
	user := GetUserID(c)
	administrators, err := GetAdministrators(c, questionnaireID)
	if err != nil {
		c.Logger().Error(err)
		return false, err
	}

	found := false
	for _, admin := range administrators {
		if admin == user || admin == "traP" {
			found = true
			break
		}
	}
	return found, nil
}
