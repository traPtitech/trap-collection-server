package router

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/repository"
)

//PostSeatHandler 席情報の更新
func PostSeatHandler(c echo.Context) error {
	seat := model.PostSeat{}
	err := c.Bind(&seat)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}

	b := repository.IsThereSeat(seat.ID)
	if seat.Status == "in" && !b {
		err = repository.InsertSeat(seat.ID)
	} else if seat.Status == "out" && b {
		err = repository.DeleteSeat(seat.ID)
	} else {
		return c.String(http.StatusInternalServerError, "status is invalid")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in updating seat status")
	}

	return c.NoContent(http.StatusOK)
}

//GetSeatHandler 席の取得
func GetSeatHandler(c echo.Context) error {
	seats, err := repository.GetSeat()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting seats")
	}
	return c.JSON(http.StatusOK, seats)
}
