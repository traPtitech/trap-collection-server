package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// NewOAuth OAuthのコンストラクタ
func NewOAuth(strURL string) (OAuth, error) {
	baseURL, err := url.Parse(strURL)
	if err != nil {
		return &oAuth{}, fmt.Errorf("Faile In Parsing URL: %w", err)
	}
	authBase := &oAuth{
		baseURL: baseURL,
	}
	return authBase, nil
}

type oAuth struct {
	baseURL *url.URL
}

func (o *oAuth) BaseURL() *url.URL {
	baseURL := *o.baseURL
	return &baseURL
}

// GetMe エンドポイントを叩いた人の取得
func (o *oAuth) GetMe(accessToken string) (*openapi.User, error) {
	path := o.BaseURL()
	path.Path += "/users/me"

	req, err := http.NewRequest("GET", path.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	httpClient := http.DefaultClient

	res, err := httpClient.Do(req)
	if err != nil {
		return &openapi.User{}, err
	}
	if res.StatusCode != 200 {
		return &openapi.User{}, fmt.Errorf("Failed In HTTP Request:(Status:%d %s)", res.StatusCode, res.Status)
	}

	user := &openapi.User{}
	err = json.NewDecoder(res.Body).Decode(user)
	if err != nil {
		return &openapi.User{}, err
	}
	return user, nil
}

// GetUsers traQのユーザー一覧の取得
func (o *oAuth) GetUsers(accessToken string) ([]*openapi.User, error) {
	path := o.BaseURL()
	path.Path += "/users"

	req, err := http.NewRequest("GET", path.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	httpClient := http.DefaultClient

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP request failed:(Status:%d %s)", res.StatusCode, res.Status)
	}

	var users []*openapi.User
	err = json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// NewLauncherAuth LauncherAuthのコンストラクタ
func NewLauncherAuth() LauncherAuth {
	newLauncherAuth := new(launcherAuth)
	return newLauncherAuth
}

type launcherAuth struct{}

// GetVersionID バージョンのIDの取得
func (*launcherAuth) GetVersionID(c echo.Context) (string, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Session: %w", err)
	}

	interfaceVersion, ok := sess.Values["versionID"]
	if !ok || interfaceVersion == nil {
		log.Println("error: unexpected no versionID")
		return "", errors.New("No VersionID")
	}
	versionID := interfaceVersion.(string)

	return versionID, nil
}

// GetProductKey プロダクトキーの取得
func (*launcherAuth) GetProductKey(c echo.Context) (string, error) {
	productKey := c.Request().Header.Get("X-Key")
	if len(productKey) == 0 {
		return "", errors.New("No Product Key")
	}

	return productKey, nil
}
