package router

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/traPtitech/trap-collection-server/storage"
)

type ioReader = io.Reader
type multipartFile = multipart.File
type sessionsSession = sessions.Session
type sessionMap = map[interface{}]interface{}

// Service serviceの構造体
type Service struct {
	baseURL *url.URL
	*Middleware
	*Game
	*OAuth2
	*Question
	*Response
	*Seat
	*User
	*Version
}

// NewService Serviceのコンストラクタ
func NewService(clientID string,clientSecret string) (Service, error) {
	str, err := storage.NewLocalStorage("../upload")
	if err != nil {
		return Service{}, fmt.Errorf("Failed In LoacalStorage Constructor: %w", err)
	}
	game := NewGame(&str)

	strBaseURL := "https://q.trap.jp/api/v3"
	authBase, err := NewOAuthBase(strBaseURL)
	if err != nil {
		return Service{}, fmt.Errorf("Failed In AuthBase Constructor: %w", err)
	}

	oAuth2:= NewOAuth2(authBase, clientID, clientSecret)
	middleware := NewMiddleware(authBase)


	api := Service{
		Middleware: &middleware,
		Game: &game,
		OAuth2: &oAuth2,
	}
	return api, nil
}
