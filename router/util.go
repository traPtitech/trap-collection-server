package router

import (
	"errors"
	"mime/multipart"
	"os"

	"github.com/gorilla/sessions"
	"github.com/traPtitech/trap-collection-server/openapi"
)

var (
	clientID     string
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

// MockAPI apiの構造体（mock）
type MockAPI struct {
	User openapi.User
	Middleware
	*MockGameApi
	*MockOauth2Api
	*MockQuestionApi
	*MockResponseApi
	*MockSeatApi
	*MockUserApi
	*MockVersionApi
}

type osFile = os.File
type multipartFile = multipart.File
type sessionsSession = sessions.Session

// InitMock mockの初期化
func InitMock() error {
	return nil
}
