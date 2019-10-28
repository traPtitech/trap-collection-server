package router

import (
	"net/http"

	"github.com/labstack/echo"
)

// GetUsersMe GET /users/me
func GetUsersMe(c echo.Context) error {
	user := c.Get("user").(string)

	return c.JSON(http.StatusOK, user)
}
