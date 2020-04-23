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
type OAuth2 struct{}

var baseURL, _ = url.Parse("https://q.trap.jp/api/1.0")

// AuthResponse 認証の返答
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Callback GET /oauth/callbackの処理部分
func (o OAuth2) Callback(code string, sessMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	interfaceCodeVerifier, ok := sessMap["codeVerifier"]
	if !ok || interfaceCodeVerifier == nil {
		return map[interface{}]interface{}{}, errors.New("CodeVerifier IS NULL")
	}
	codeVerifier := interfaceCodeVerifier.(string)
	res, err := getAccessToken(code, codeVerifier)
	if err != nil {
		return map[interface{}]interface{}{}, fmt.Errorf("Failed In Getting AccessToken:%w", err)
	}

	sessMap["accessToken"] = res.AccessToken
	sessMap["refreshToken"] = res.RefreshToken

	user, err := getMe(res.AccessToken)
	if err != nil {
		return map[interface{}]interface{}{}, fmt.Errorf("Failed In Getting Me: %w", err)
	}

	sessMap["userID"] = user.UserId
	sessMap["userName"] = user.Name

	return sessMap, nil
}

// GetGenerateCode POST /oauth/generate/codeの処理部分
func (o OAuth2) GetGenerateCode() (openapi.InlineResponse200, map[interface{}]interface{}, error) {
	pkceParams := openapi.InlineResponse200{}

	pkceParams.ResponseType = "code"

	pkceParams.ClientId = clientID

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

// PostLogout POST /oauth/logoutの処理部分
func (o OAuth2) PostLogout(sessMap map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	interfaceAccessToken, ok := sessMap["accessToken"]
	if !ok || interfaceAccessToken == nil {
		return map[interface{}]interface{}{}, errors.New("AccessToken IS NULL")
	}
	accessToken := interfaceAccessToken.(string)

	path := *baseURL
	path.Path += "/oauth2/revoke"
	form := url.Values{}
	form.Set("token", accessToken)
	reqBody := strings.NewReader(form.Encode())
	req, err := http.NewRequest("POST", path.String(), reqBody)
	if err != nil {
		return map[interface{}]interface{}{}, fmt.Errorf("Failed In Making HTTP Request:%w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return map[interface{}]interface{}{}, fmt.Errorf("Failed In HTTP Request:%w", err)
	}
	if res.StatusCode != 200 {
		return map[interface{}]interface{}{}, fmt.Errorf("Failed In Getting Access Token:(Status:%d %s)", res.StatusCode, res.Status)
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

func getAccessToken(code string, codeVerifier string) (AuthResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	reqBody := strings.NewReader(form.Encode())
	path := *baseURL
	path.Path += "/oauth2/token"
	req, err := http.NewRequest("POST", path.String(), reqBody)
	if err != nil {
		return AuthResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return AuthResponse{}, err
	}
	if res.StatusCode != 200 {
		return AuthResponse{}, fmt.Errorf("Failed In Getting Access Token:(Status:%d %s)", res.StatusCode, res.Status)
	}
	var authRes AuthResponse
	err = json.NewDecoder(res.Body).Decode(&authRes)
	if err != nil {
		return AuthResponse{}, err
	}
	return authRes, nil
}

func getMe(accessToken string) (openapi.User, error) {
	path := *baseURL
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
