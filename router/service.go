package router

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
	"github.com/traPtitech/trap-collection-server/session"
	"github.com/traPtitech/trap-collection-server/storage"
)

type ioReader = io.Reader
type multipartFile = multipart.File

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
func NewAPI(sess session.Session, env string, clientID string,clientSecret string) (*openapi.Api, error) {
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
	oauth, err := base.NewOAuth(strBaseURL)
	if err != nil {
		return &openapi.Api{}, fmt.Errorf("Failed In OAuth Constructor: %w", err)
	}

	oAuth2:= NewOAuth2(sess, oauth, clientID, clientSecret)
	middleware := NewMiddleware(oauth)

	api := &openapi.Api{
		Middleware: middleware,
		GameApi: game,
		Oauth2Api: oAuth2,
	}
	return api, nil
}
