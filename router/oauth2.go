package router

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// OAuth2 oauthの構造体
type OAuth2 struct {
	openapi.Oauth2Api
}

var baseURL, _ = url.Parse("https://q.trap.jp/api/1.0")

// AuthResponse 認証の返答
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// CallbackHandler GET /oauth/callbackの処理部分
func (o OAuth2) CallbackHandler(code string, c echo.Context) (echo.Context, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return c, fmt.Errorf("Failed In Getting Session:%w", err)
	}

	codeVerifier := sess.Values["codeVerifier"].(string)
	res, err := getAccessToken(code, codeVerifier)
	if err != nil {
		return c, fmt.Errorf("Failed In Getting AccessToken:%w", err)
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   res.ExpiresIn * 1000,
		HttpOnly: true,
	}

	sess.Values["accessToken"] = res.AccessToken
	sess.Values["refreshToken"] = res.RefreshToken

	user, err := getMe(res.AccessToken)
	if err != nil {
		return c, fmt.Errorf("Failed In Getting Me: %w", err)
	}

	sess.Values["id"] = user.UserId
	sess.Values["name"] = user.Name
	sess.Save(c.Request(), c.Response())

	return c, nil
}

// GetGenerateCodeHandler POST /oauth/generate/codeの処理部分
func (o OAuth2) GetGenerateCodeHandler(c echo.Context) (openapi.InlineResponse200, echo.Context, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return openapi.InlineResponse200{}, c, fmt.Errorf("Failed In Getting Session:%w", err)
	}

	pkceParams := openapi.InlineResponse200{}

	pkceParams.ResponseType = "code"

	pkceParams.ClientId = clientID

	bytesCodeVerifier := randBytes(43)
	codeVerifier := string(bytesCodeVerifier)
	bytesCodeChallenge := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64url.Encode(bytesCodeChallenge[:])
	pkceParams.CodeChallenge = codeChallenge
	sess.Values["codeVerifier"] = codeVerifier

	pkceParams.CodeChallengeMethod = "S256"

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 1000,
		HttpOnly: true,
	}

	sess.Save(c.Request(), c.Response())

	return pkceParams, c, nil
}

// PostLogoutHandler POST /oauth/logoutの処理部分
func (o OAuth2) PostLogoutHandler(c echo.Context) (echo.Context, error) {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return c, fmt.Errorf("Failed In Getting Session:%w", err)
	}

	accessToken := sess.Values["accessToken"].(string)
	path := *baseURL
	path.Path += "/oauth2/revoke"
	form := url.Values{}
	form.Set("token",accessToken)
	reqBody := strings.NewReader(form.Encode())
	req, err := http.NewRequest("POST", path.String(), reqBody)
	if err != nil {
		return c, fmt.Errorf("Failed In Making HTTP Request:%w",err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return c, fmt.Errorf("Failed In HTTP Request:%w",err)
	}
	if res.StatusCode != 200 {
		return c, fmt.Errorf("Failed In Getting Access Token:(Status:%d %s)", res.StatusCode, res.Status)
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
	}

	sess.Values["accessToken"] = nil
	sess.Values["refreshToken"] = nil
	sess.Values["id"] = nil
	sess.Values["name"] = nil

	sess.Save(c.Request(), c.Response())

	return c, nil
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
