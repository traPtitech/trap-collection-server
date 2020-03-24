package router

import (
	echo "github.com/labstack/echo/v4"
)

//SetupRouting ルーティング
func SetupRouting(e *echo.Echo, client Traq) {
	apiNoAuth := e.Group("/api")
	{
		apiNoAuth.GET("/callback", CallbackHandler)
	}
	api := e.Group("/api", client.MiddlewareAuthUser)
	{
		api.GET("/users/me", GetMeHandler(client))
		api.POST("/logout", PostLogoutHandler)
	}
}
