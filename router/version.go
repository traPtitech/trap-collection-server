package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
)

// GetCheckListHandler GET /version/check/{launcherVersionID}のハンドラー
func GetCheckListHandler(c echo.Context) error {
	return c.JSON(http.StatusOK,[]string{})
}

// GetVersionHandler GET /version/{launcherVersionID}のハンドラー
func GetVersionHandler(c echo.Context) error {
	launcherVersionID,err := strconv.Atoi(c.Param("launcherVersionID"))
	if err != nil {
		return c.String(http.StatusBadRequest,fmt.Errorf("Failed In Comverting Launcher Version ID:%w",err).Error())
	}
	launcherVersion,err := model.GetLauncherVersionByID(uint(launcherVersionID))
	if err != nil {
		return c.String(http.StatusInternalServerError,fmt.Errorf("Failed In Getting Launcher Version ID:%w",err).Error())
	}
	return c.JSON(http.StatusOK,launcherVersion)
}