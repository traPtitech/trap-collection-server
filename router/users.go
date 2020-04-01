package router

import (
	"fmt"
	"net/http"

	echo "github.com/labstack/echo/v4"
)

// GetMeHandler GET /users/meのハンドラー
func GetMeHandler(client Traq) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := client.GetMe(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Me:%w", err).Error())
		}
		return c.JSON(http.StatusOK, user)
	}
}
