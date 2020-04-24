package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/traPtitech/trap-collection-server/openapi"
)

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
	oAuth2, err := NewOAuth2(strBaseURL, clientID, clientSecret)
	if err != nil {
		return Service{}, fmt.Errorf("Failed In OAuth2 Constructor: %w", err)
	}
	middleware, err := NewMiddleware(strBaseURL)
	if err != nil {
		return Service{}, fmt.Errorf("Failed In Middleware Constructor: %w", err)
	}
	api := Service{
		Middleware: &middleware,
		OAuth2: &oAuth2,
	}
	return api, nil
}

func (s *Service) getMe(accessToken string) (openapi.User, error) {
	path := *s.baseURL
	path.Path += "/users/me"
	req, err := http.NewRequest("GET", path.String(), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return openapi.User{}, err
	}
	if res.StatusCode != 200 {
		return openapi.User{}, fmt.Errorf("Failed In HTTP Request:(Status:%d %s)", res.StatusCode, res.Status)
	}
	var user openapi.User
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return openapi.User{}, err
	}
	return user, nil
}
