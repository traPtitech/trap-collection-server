package router

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"

	"github.com/gorilla/sessions"
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
	strBaseURL := "https://q.trap.jp/api/1.0"
	authBase, err := NewAuthBase(strBaseURL)
	if err != nil {
		return Service{}, fmt.Errorf("Failed In AuthBase Constructor: %w", err)
	}
	oAuth2:= NewOAuth2(authBase, clientID, clientSecret)
	middleware := NewMiddleware(authBase)
	api := Service{
		Middleware: &middleware,
		OAuth2: &oAuth2,
	}
	return api, nil
}
