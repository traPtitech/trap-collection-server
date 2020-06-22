package router

import (
	"errors"
	"fmt"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// User userの構造体
type User struct {
	openapi.UserApi
}

// GetMe GET /users/meの処理部分
func (*User) GetMe(c echo.Context) (*openapi.User, error) {
	sess,err := session.Get("sessions", c)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Session: %w", err)
	}

	userID, ok := sess.Values["userID"]
	if !ok || userID == nil {
		return nil, errors.New("userID IS NULL")
	}

	userName, ok := sess.Values["userName"]
	if !ok || userName == nil {
		return nil, errors.New("userName IS NULL")
	}

	user := new(openapi.User)
	user.Id = userID.(string)
	user.Name = userName.(string)

	return user, nil
}
