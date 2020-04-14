package router

import (
	"fmt"

	echo "github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// User userの構造体
type User struct {
	openapi.UserApi
}

// GetMe GET /users/meの処理部分
func (u User)GetMe(c echo.Context) (openapi.User, echo.Context, error) {
	user, err := GetMe(c)
	if err != nil {
		return openapi.User{}, c, fmt.Errorf("Failed In Getting Me:%w", err)
	}
	return user, c, nil
}
