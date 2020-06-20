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
	sessMap := sess.Values

	userID, ok := sessMap["userID"]
	if !ok || userID == nil {
		return &openapi.User{}, errors.New("userID IS NULL")
	}

	userName, ok := sessMap["userName"]
	if !ok || userName == nil {
		return &openapi.User{}, errors.New("userName IS NULL")
	}

	user := &openapi.User{
		Id: userID.(string),
		Name:   userName.(string),
	}

	return user, nil
}
