package router

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
)

//GetAdminsHandler 管理者の一覧を取得する関数
func GetAdminsHandler(c echo.Context) error {
	admins,err := repository.GetAdministrators(c)
	if err!=nil{
		return c.String(http.StatusInternalServerError, "something wrong in getting admins")
	}
	return c.JSON(http.StatusOK, &admins)
}

//PushAdminsHandler 管理者を追加する関数
func PushAdminsHandler(c echo.Context) error {
	type AddAdmins struct{
		Admins []string `json:"admins,omitempty"`
	}
	admins := AddAdmins{}
	c.Bind(&admins)

	err := repository.InsertAdministrators(c,admins.Admins)
	if err!=nil{
		return c.String(http.StatusInternalServerError, "something wrong in inserting admins")
	}
	return nil
}

//DeleteAdminHandler 管理者を削除する関数
func DeleteAdminHandler(c echo.Context) error {
	type Admin struct{
		ID string `json:"id,omitempty"`
	}
	admin := Admin{}
	c.Bind(&admin)

	err := repository.DeleteAdministrators(c,admin.ID)
	if err!=nil{
		return c.String(http.StatusInternalServerError, "something wrong in deleting admins")
	}
	return nil
}