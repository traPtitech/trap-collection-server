package router

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/gorilla/sessions"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/storage"
)

type ioReader = io.Reader
type multipartFile = multipart.File
type sessionsSession = sessions.Session
type sessionMap = map[interface{}]interface{}

// Service serviceの構造体
type Service struct {
	*Middleware
	*Game
	*OAuth2
	*Question
	*Response
	*Seat
	*User
	*Version
}

// NewAPI Apiのコンストラクタ
func NewAPI(env string, clientID string,clientSecret string) (*openapi.Api, error) {
	var str storage.Storage
	if env == "development" || env == "mock" {
		localStr, err := storage.NewLocalStorage("./upload")
		if err != nil {
			return &openapi.Api{}, fmt.Errorf("Failed In LoacalStorage Constructor: %w", err)
		}
		str = localStr
	} else {
		swiftStr, err := storage.NewSwiftStorage(os.Getenv("container"))
		if err != nil {
			return &openapi.Api{}, fmt.Errorf("Failed In Swift Storage Constructor: %w", err)
		}
		str = swiftStr
	}
	game := NewGame(str)

	strBaseURL := "https://q.trap.jp/api/v3"
	authBase, err := NewOAuthBase(strBaseURL)
	if err != nil {
		return &openapi.Api{}, fmt.Errorf("Failed In AuthBase Constructor: %w", err)
	}

	oAuth2:= NewOAuth2(authBase, clientID, clientSecret)
	middleware := NewMiddleware(authBase)

	api := &openapi.Api{
		Middleware: middleware,
		GameApi: game,
		Oauth2Api: oAuth2,
	}
	return api, nil
}
