package router

import (
	"errors"
	"fmt"
	"os"

	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
)

var (
	clientID string
	clientSecret string
)

// API apiの構造体
type API struct {
	Middleware
	Game
	OAuth2
	Question
	Response
	Seat
	User
	Version
}

// InitRouter router内で使う環境変数の初期化
func InitRouter() error {
	clientID = os.Getenv("CLIENT_ID")
	if len(clientID) == 0 {
		return errors.New("ENV CLIENT_ID IS NULL")
	}
	clientSecret = os.Getenv("CLIENT_SECRET")
	if len(clientSecret) == 0 {
		return errors.New("ENV CLIENT_SECRET IS NULL")
	}
	return nil
}

// GetMe sessionからuserのID、名前を取得
func GetMe(c echo.Context) (openapi.User, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return openapi.User{}, fmt.Errorf("Failed In Getting Session:%w", err)
	}
	id := sess.Values["id"].(string)
	name := sess.Values["name"].(string)
	if len(id) == 0 || len(name) == 0 {
		accessToken := sess.Values["accessToken"].(string)
		if len(accessToken) == 0 {
			return openapi.User{}, errors.New("AccessToken Is Null")
		}
		user, err := getMe(accessToken)
		if err != nil {
			return openapi.User{}, fmt.Errorf("Failed In Getting Me:%w", err)
		}
		return user, nil
	}
	return openapi.User{UserId: id, Name: name}, nil
}
