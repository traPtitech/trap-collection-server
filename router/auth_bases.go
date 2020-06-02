package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/traPtitech/trap-collection-server/openapi"
)

// OAuthBase OAuthの認証の基本部分の構造体
type OAuthBase struct {
	baseURL *url.URL
}

// NewOAuthBase AuthBaseのコンストラクタ
func NewOAuthBase(strURL string) (*OAuthBase,error) {
	baseURL, err := url.Parse(strURL)
	if err != nil {
		return &OAuthBase{}, fmt.Errorf("Faile In Parsing URL: %w", err)
	}
	authBase := &OAuthBase{
		baseURL: baseURL,
	}
	return authBase, nil
}

func (a *OAuthBase) getMe(accessToken string) (*openapi.User, error) {
	path := *a.baseURL
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

// LauncherAuthBase ランチャーの認証の基本部分の構造体
type LauncherAuthBase struct {}

func (*LauncherAuthBase) getVersionID(sess sessionMap) (uint, error) {
	interfaceVersion, ok := sess["versionID"]
	if !ok || interfaceVersion == nil {
		log.Println("error: unexpected no versionID")
		return 0, errors.New("No VersionID")
	}
	versionID := interfaceVersion.(uint)
	return versionID, nil
}

func (*LauncherAuthBase) getProductKey(sess sessionMap) (string, error) {
	interfaceProductKey, ok := sess["productKey"]
	if !ok || interfaceProductKey == nil {
		log.Println("error: unexpected no productKey")
		return "", errors.New("No ProductKey")
	}
	productKey := interfaceProductKey.(string)
	return productKey, nil
}