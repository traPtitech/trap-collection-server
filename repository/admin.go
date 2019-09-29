package repository

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
)

//GetAdministrators 管理者の取得
func GetAdministrators(c echo.Context) ([]model.AdminList, error) {
	administrators := []model.AdminList{}
	if err := Db.Select(&administrators, "SELECT user_traqid FROM administrators"); err != nil {
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return administrators, nil
}

//InsertAdministrators 管理者の追加
func InsertAdministrators(c echo.Context, administrators []string) error {
	//N+1 そのうち解消した方がいい
	for _, v := range administrators {
		if _, err := Db.Exec(
			"INSERT INTO administrators user_traqid VALUES ?",
			v); err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	}
	return nil
}

//DeleteAdministrators 管理者の削除
func DeleteAdministrators(c echo.Context, id string) error {
	if _, err := Db.Exec(
		"DELETE from administrators WHERE user_traqid = ?",
		id); err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}

//CheckAdmin 自分がadminなら(true, nil)
func CheckAdmin(c echo.Context) (bool, error) {
	user := GetUserID(c)
	var id string
	err := Db.Get(&id, "SELECT user_traqid FROM administrators WHERE user_traqid = ?", user)
	if err != nil {
		return false, nil
	}
	return true, nil
}
