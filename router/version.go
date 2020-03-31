package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetCheckListHandler チェックリストの取得
func GetCheckListHandler(c echo.Context) error {
	return c.JSON(http.StatusOK,[]string{})
}