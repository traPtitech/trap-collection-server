package router

import (
	echo "github.com/labstack/echo/v4"
)

//SetupRouting ルーティング
func SetupRouting(e *echo.Echo, client Traq) {
	apiNoAuth := e.Group("/api")
	{
		apiOAuth := apiNoAuth.Group("/oauth2")
		{
			apiOAuth.GET("/callback", CallbackHandler)
			apiOAuth.POST("/generate/code", GetGenerateCodeHandler)
		}
	}
	api := e.Group("/api", client.MiddlewareAuthUser)
	{
		apiUsers := api.Group("/users")
		{
			apiUsers.GET("/me", GetMeHandler(client))
		}
		apiOAuth := api.Group("/oauth2")
		{
			apiOAuth.POST("/logout", PostLogoutHandler)
		}
	}
}
