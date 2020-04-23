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
func (u User) GetMe(sessMap map[interface{}]interface{}) (openapi.User, map[interface{}]interface{}, error) {
	userID, ok := sessMap["userID"]
	if !ok || userID == nil {
		return openapi.User{}, map[interface{}]interface{}{}, errors.New("userID IS NULL")
	}

	userName, ok := sessMap["userName"]
	if !ok || userName == nil {
		return openapi.User{}, map[interface{}]interface{}{}, errors.New("userName IS NULL")
	}

	user := openapi.User{
		UserId: userID.(string),
		Name:   userName.(string),
	}

	return user, map[interface{}]interface{}{}, nil
}
