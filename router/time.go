package router

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/repository"
)

//PostTimeHandler 時間の追加
func PostTimeHandler(c echo.Context) error {
	time := model.Time{}
	err := c.Bind(&time)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}

	for _, v := range time.List {
		err := repository.InsertTime(time.VersionID, v.GameID, v.StartTime, v.EndTime)
		if err != nil {
			return c.String(http.StatusInternalServerError, "something wrong in inserting time")
		}
	}
	return c.NoContent(http.StatusOK)
}
