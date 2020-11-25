package router

import (
	"errors"
	"fmt"
	"log"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
)

// User userの構造体
type User struct {
	openapi.UserApi
	base.OAuth
}

func newUser(oauth base.OAuth) *User {
	user := &User{
		OAuth: oauth,
	}

	return user
}

// GetMe GET /users/meの処理部分
func (*User) GetMe(c echo.Context) (*openapi.User, error) {
	sess, err := session.Get("sessions", c)
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

// GetUsers GET /usersの処理部分
func (u *User) GetUsers(c echo.Context) ([]*openapi.User, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting Session:%w", err)
	}

	interfaceAccessToken, ok := sess.Values["accessToken"]
	if !ok {
		log.Println("unexpected getting access token error")
		return nil, errors.New("Failed In Getting Access Token")
	}

	accessToken, ok := interfaceAccessToken.(string)
	if !ok {
		log.Println("unexpected parsing access token error")
		return nil, errors.New("Failed In Parsing Access Token")
	}

	users, err := u.OAuth.GetUsers(accessToken)
	if err != nil {
		return nil, fmt.Errorf("Failed In Getting traQ Users: %w", err)
	}

	return users, nil
}
