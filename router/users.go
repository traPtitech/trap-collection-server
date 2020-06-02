package router

import (
	"errors"

	"github.com/traPtitech/trap-collection-server/openapi"
)

// User userの構造体
type User struct {
	openapi.UserApi
}

// GetMe GET /users/meの処理部分
func (*User) GetMe(sessMap sessionMap) (*openapi.User, sessionMap, error) {
	userID, ok := sessMap["userID"]
	if !ok || userID == nil {
		return &openapi.User{}, sessionMap{}, errors.New("userID IS NULL")
	}

	userName, ok := sessMap["userName"]
	if !ok || userName == nil {
		return &openapi.User{}, sessionMap{}, errors.New("userName IS NULL")
	}

	user := &openapi.User{
		Id: userID.(string),
		Name:   userName.(string),
	}

	return user, sessionMap{}, nil
}
