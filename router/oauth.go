package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
)

var baseURL = "https://q.trap.jp/api/1.0"

type authResponse struct {
	accessToken  string `json:"access_token"`
	expiresIn    int    `json:"expires_in"`
	refreshToken string `json:"refresh_token"`
}

// CallbackHandler OAuthのコールバック
func CallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")

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
		MaxAge:   res.expiresIn,
		HttpOnly: true,
	}
	sess.Values["accessToken"] = res.accessToken
	sess.Values["refreshToken"] = res.refreshToken
	user, err := getMe(res.accessToken)
	sess.Values["id"] = user.ID
	sess.Values["name"] = user.Name
	sess.Save(c.Request(), c.Response())
	return c.NoContent(http.StatusOK)
}

func getAccessToken(code string, codeVerifier string) (authResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/oauth2/token", nil)
	if err != nil {
		return authResponse{}, err
	}
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", clientID)
	form.Add("code", code)
	form.Add("code_verifier", codeVerifier)

	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return authResponse{}, errors.New(err.Error())
	}
	if res.StatusCode != 200 {
		return authResponse{}, errors.New("認証に失敗しました")
	}
	body, _ := ioutil.ReadAll(res.Body)
	authRes := authResponse{}
	err = json.Unmarshal(body, &authRes)
	if err != nil {
		return authResponse{}, err
	}
	return authRes, nil
}

func getMe(accessToken string) (User, error) {
	req, err := http.NewRequest("GET", baseURL+"/users/me", nil)
	req.Header.Set("Authorization", accessToken)
	httpClient := http.DefaultClient
	res, err := httpClient.Do(req)
	if err != nil {
		return User{}, err
	}
	if res.StatusCode != 200 {
		return User{}, fmt.Errorf("Failed In HTTP Request:%d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	user := User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
