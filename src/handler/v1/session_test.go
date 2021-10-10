package v1

import "github.com/labstack/echo/v4"

// テスト用のレスポンスのSet-CookieヘッダーをCookieヘッダーに移す関数
func setCookieHeader(c echo.Context) {
	cookie := c.Response().Header().Get("Set-Cookie")
	c.Response().Header().Del("Set-Cookie")
	c.Request().Header.Set("Cookie", cookie)
}
