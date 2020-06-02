package router

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// OAuth2 oauthの構造体
type OAuth2 struct {
	*OAuthBase
	clientID     string
	clientSecret string
}

// NewOAuth2 OAuth2のコンストラクタ
func NewOAuth2(authBase *OAuthBase, clientID string, clientSecret string) *OAuth2 {
	oAuth2 := &OAuth2{
		OAuthBase: authBase,
		clientID: clientID,
		clientSecret: clientSecret,
	}
	return oAuth2
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Callback GET /oauth2/callbackの処理部分
func (o *OAuth2) Callback(code string, sessMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	interfaceCodeVerifier,ok := sessMap["codeVerifier"]
	if !ok || interfaceCodeVerifier == nil {
		return sessionMap{}, errors.New("CodeVerifier IS NULL")
	}
	codeVerifier := interfaceCodeVerifier.(string)
	res, err := o.getAccessToken(code, codeVerifier)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In Getting AccessToken:%w", err)
	}

	sessMap["accessToken"] = res.AccessToken
	sessMap["refreshToken"] = res.RefreshToken

	user, err := o.getMe(res.AccessToken)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In Getting Me: %w", err)
	}

	sessMap["userID"] = user.Id
	sessMap["userName"] = user.Name

	return sessMap, nil
}

// GetGenerateCode POST /oauth2/generate/codeの処理部分
func (o *OAuth2) GetGenerateCode() (*openapi.InlineResponse200, map[interface{}]interface{}, error) {
	pkceParams := &openapi.InlineResponse200{}

	pkceParams.ResponseType = "code"

	pkceParams.ClientId = o.clientID

	bytesCodeVerifier := randBytes(43)
	codeVerifier := string(bytesCodeVerifier)
	bytesCodeChallenge := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64url.Encode(bytesCodeChallenge[:])
	pkceParams.CodeChallenge = codeChallenge

	sessMap := make(map[interface{}]interface{})
	sessMap["codeVerifier"] = codeVerifier

	pkceParams.CodeChallengeMethod = "S256"

	return pkceParams, sessMap, nil
}

// PostLogout POST /oauth2/logoutの処理部分
func (o *OAuth2) PostLogout(sessMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	interfaceAccessToken,ok := sessMap["accessToken"]
	if !ok || interfaceAccessToken == nil {
		return sessionMap{}, errors.New("AccessToken IS NULL")
	}
	accessToken := interfaceAccessToken.(string)

	path := *o.baseURL
	path.Path += "/oauth2/revoke"
	form := url.Values{}
	form.Set("token",accessToken)
	reqBody := strings.NewReader(form.Encode())
	req, err := http.NewRequest("POST", path.String(), reqBody)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In Making HTTP Request:%w",err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return sessionMap{}, fmt.Errorf("Failed In HTTP Request:%w",err)
	}
	if res.StatusCode != 200 {
		return sessionMap{}, fmt.Errorf("Failed In Getting Access Token:(Status:%d %s)", res.StatusCode, res.Status)
	}

	sessMap["accessToken"] = nil
	sessMap["refreshToken"] = nil
	sessMap["userID"] = nil
	sessMap["userName"] = nil

	return sessMap, nil
}

var randSrc = rand.NewSource(time.Now().UnixNano())

const (
	letters       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func randBytes(n int) []byte {
	b := make([]byte, n)
	cache, remain := randSrc.Int63(), letterIdxMax
	for i := n - 1; i >= 0; {
		if remain == 0 {
			cache, remain = randSrc.Int63(), letterIdxMax
		}
		idx := int(cache & letterIdxMask)
		if idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return b
}

func (o *OAuth2) getAccessToken(code string, codeVerifier string) (*authResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", o.clientID)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	reqBody := strings.NewReader(form.Encode())
	path := *o.baseURL
	path.Path += "/oauth2/token"
	req, err := http.NewRequest("POST", path.String(), reqBody)
	if err != nil {
		return &authResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return &authResponse{}, err
	}
	if res.StatusCode != 200 {
		return &authResponse{}, fmt.Errorf("Failed In Getting Access Token:(Status:%d %s)", res.StatusCode, res.Status)
	}
	var authRes *authResponse
	err = json.NewDecoder(res.Body).Decode(authRes)
	if err != nil {
		return &authResponse{}, err
	}
	return authRes, nil
}
