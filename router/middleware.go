package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo"

	"github.com/traPtitech/booQ/model"
)

var baseURL = "https://q.trap.jp/api/1.0"

// Traq traQに接続する用のclient
type Traq interface {
	GetUsersMe(c echo.Context) (echo.Context, error)
	MiddlewareAuthUser(next echo.HandlerFunc) echo.HandlerFunc
}

// TraqClient 本番用のclient
type TraqClient struct {
	Traq
}

// MockTraqClient テスト用のモックclient
type MockTraqClient struct {
	Traq
	MockGetUsersMe func(c echo.Context) (echo.Context, error)
}

// GetUsersMe 本番用のGetUsersMe
func (client *TraqClient) GetUsersMe(c echo.Context) (echo.Context, error) {
	tokenCookie, err := c.Cookie("token")
	token := ""
	if err != nil {
		token = c.Request().Header.Get("Authorization")
	} else {
		token = strings.Replace(strings.Replace(tokenCookie.String(), "token=\"", "", 1), "\"", "", 1)
		fmt.Println(token)
		fmt.Println("Cookie")
		if token == "" {
			token = c.Request().Header.Get("Authorization")
		}
	}

	if token == "" {
		return c, errors.New("認証に失敗しました(Headerに必要な情報が存在しません)")
	}
	req, _ := http.NewRequest("GET", baseURL+"/users/me", nil)
	req.Header.Set("Authorization", token)
	httpClient := new(http.Client)
	res, _ := httpClient.Do(req)
	if res.StatusCode != 200 {
		fmt.Println(token)
		return c, errors.New("認証に失敗しました")
	}
	body, _ := ioutil.ReadAll(res.Body)
	traqUser := model.User{}
	_ = json.Unmarshal(body, &traqUser)
	c.Set("user", traqUser.Name)
	c.SetCookie(&http.Cookie{Name: "token", Value: token})

	return c, nil
}

// GetUsersMe テスト用のGetUsersMe
func (client *MockTraqClient) GetUsersMe(c echo.Context) (echo.Context, error) {
	return client.MockGetUsersMe(c)
}

// MiddlewareAuthUser APIにアクセスしたユーザーの情報をセットする
func (client *TraqClient) MiddlewareAuthUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c, err := client.GetUsersMe(c)
		if err != nil {
			return c.String(http.StatusUnauthorized, err.Error())
		}
		return next(c)
	}
}

// MiddlewareAuthUser APIにアクセスしたユーザーの情報をセットする
func (client *MockTraqClient) MiddlewareAuthUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c, err := client.GetUsersMe(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}
		return next(c)
	}
}
