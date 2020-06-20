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
func NewOAuth(strURL string) (OAuth,error) {
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
	return o.baseURL
}

// GetMe エンドポイントを叩いた人の取得
func (o *oAuth) GetMe(accessToken string) (*openapi.User, error) {
	path := *o.baseURL
	path.Path += "/users/me"
	req, err := http.NewRequest("GET", path.String(), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return &openapi.User{}, err
	}
	if res.StatusCode != 200 {
		return &openapi.User{}, fmt.Errorf("Failed In HTTP Request:(Status:%d %s)", res.StatusCode, res.Status)
	}
	var user *openapi.User
	err = json.NewDecoder(res.Body).Decode(user)
	if err != nil {
		return &openapi.User{}, err
	}
	return user, nil
}

// NewLauncherAuth LauncherAuthのコンストラクタ
func NewLauncherAuth() LauncherAuth {
	newLauncherAuth := new(launcherAuth)
	return newLauncherAuth
}

type launcherAuth struct {}

// GetVersionID バージョンのIDの取得
func (*launcherAuth) GetVersionID(c echo.Context) (uint, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return 0, fmt.Errorf("Failed In Getting Session: %w", err)
	}

	interfaceVersion, ok := sess.Values["versionID"]
	if !ok || interfaceVersion == nil {
		log.Println("error: unexpected no versionID")
		return 0, errors.New("No VersionID")
	}
	versionID := interfaceVersion.(uint)

	return versionID, nil
}

// GetProductKey プロダクトキーの取得
func (*launcherAuth) GetProductKey(c echo.Context) (string, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return "", fmt.Errorf("Failed In Getting Session: %w", err)
	}

	interfaceProductKey, ok := sess.Values["productKey"]
	if !ok || interfaceProductKey == nil {
		log.Println("error: unexpected no productKey")
		return "", errors.New("No ProductKey")
	}
	productKey := interfaceProductKey.(string)
	return productKey, nil
}