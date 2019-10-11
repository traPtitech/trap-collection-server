package router

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/traPtitech/anke-to/model"
)

// GetUsersMe GET /users/me
func GetUsersMe(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"traqID": model.GetUserID(c),
	})
}