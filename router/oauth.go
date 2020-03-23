package router

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
)

var baseURL, _ = url.Parse("https://q.trap.jp/api/1.0")

// AuthResponse 認証の返答
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// CallbackHandler OAuthのコールバック
func CallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	if len(code) == 0 {
		return c.String(http.StatusBadRequest, "Code Is Null")
	}

	sess, err := session.Get("sessions", c)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
	}
	if state != sess.Values["state"] {
		return c.String(http.StatusUnauthorized, "Failed In Getting State")
	}
	codeVerifier := sess.Values["codeVerifier"].(string)
	res, err := getAccessToken(code, codeVerifier)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting AccessToken:%w", err).Error())
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
		return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Me: %w", err).Error())
	}
	sess.Values["id"] = user.ID
	sess.Values["name"] = user.Name
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/api/users/me") //今はエンドポイントが/api/users/meしかないためこうしているが、最終的には変更する
}

func redirectAuth(c echo.Context) error {
	sess, err := session.Get("sessions", c)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Getting Session:%w", err).Error())
	}

	u := *baseURL
	u.Path = baseURL.Path + "/oauth2/authorize"
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Parsing URL:%w", err).Error())
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("Failed In Parsing Query:%w", err).Error())
	}

	q.Add("response_type", "code")

	q.Add("client_id", clientID)

	state := string(randBytes(10))
	q.Add("state", state)
	sess.Values["state"] = state

	bytesCodeVerifier := randBytes(43)
	codeVerifier := string(bytesCodeVerifier)
	bytesCodeChallenge := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64url.Encode(bytesCodeChallenge[:])
	log.Println(codeChallenge)
	q.Add("code_challenge", codeChallenge)
	sess.Values["codeVerifier"] = codeVerifier

	q.Add("code_challenge_method", "S256")

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 1000,
		HttpOnly: true,
	}

	sess.Save(c.Request(), c.Response())

	u.RawQuery = q.Encode()
	url := u.String()
	return c.Redirect(http.StatusFound, url)
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
	path.Path = path.Path + "/oauth2/token"
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
	body, _ := ioutil.ReadAll(res.Body)
	authRes := AuthResponse{}
	err = json.Unmarshal(body, &authRes)
	if err != nil {
		return AuthResponse{}, err
	}
	return authRes, nil
}

func getMe(accessToken string) (User, error) {
	path := *baseURL
	path.Path = path.Path + "/users/me"
	req, err := http.NewRequest("GET", path.String(), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return User{}, err
	}
	if res.StatusCode != 200 {
		return User{}, fmt.Errorf("Failed In HTTP Request::(Status:%d %s)", res.StatusCode, res.Status)
	}
	body, err := ioutil.ReadAll(res.Body)
	user := User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
