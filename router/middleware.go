package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
)

// User ユーザーの構造体
type User struct {
	ID   string `json:"userId,omitempty"`
	Name string `json:"name,omitempty"`
}

// Traq traQのOAuthのClient
type Traq interface {
	GetMe(c echo.Context) (User, error)
	MiddlewareAuthUser(next echo.HandlerFunc) echo.HandlerFunc
}

// TraqClient 本番用のclient
type TraqClient struct {
	Traq
}

// MockTraqClient テスト用のモックclient
type MockTraqClient struct {
	Traq
	User User
}

// GetMe 本番用のGetMe
func (client *TraqClient) GetMe(c echo.Context) (User, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return User{}, fmt.Errorf("Failed In Getting Session:%w", err)
	}
	id := sess.Values["id"].(string)
	name := sess.Values["name"].(string)
	if len(id) == 0 || len(name) == 0 {
		accessToken := sess.Values["accessToken"].(string)
		if len(accessToken) == 0 {
			return User{}, errors.New("AccessToken Is Null")
		}
		user, err := getMe(accessToken)
		if err != nil {
			return User{}, fmt.Errorf("Failed In Getting Me:%w", err)
		}
		return user, nil
	}
	return User{ID: id, Name: name}, nil
}

// GetMe テスト用のGetMe
func (client *MockTraqClient) GetMe(c echo.Context) (User, error) {
	return client.User, nil
}

// MiddlewareAuthUser 本番用のAPIにアクセスしたユーザーを認証するミドルウェア
func (client *TraqClient) MiddlewareAuthUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("sessions", c)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
		}
		accessToken := sess.Values["accessToken"]
		if accessToken == nil {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}

// MiddlewareAuthUser テスト用のミドルウェア
func (client *MockTraqClient) MiddlewareAuthUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

// MiddlewareAuthLancher ランチャーの認証用のミドルウェア
func MiddlewareAuthLancher(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
