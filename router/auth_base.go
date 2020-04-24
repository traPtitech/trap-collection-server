package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/traPtitech/trap-collection-server/openapi"
)

// AuthBase 認証の基本部分の構造体
type AuthBase struct {
	baseURL *url.URL
}

// NewAuthBase AuthBaseのコンストラクタ
func NewAuthBase(strURL string) (AuthBase,error) {
	baseURL, err := url.Parse(strURL)
	if err != nil {
		return AuthBase{}, fmt.Errorf("Faile In Parsing URL: %w", err)
	}
	authBase := AuthBase{
		baseURL: baseURL,
	}
	return authBase, nil
}

func (a *AuthBase) getMe(accessToken string) (openapi.User, error) {
	path := *a.baseURL
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