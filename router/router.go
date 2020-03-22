package router

import (
	echo "github.com/labstack/echo/v4"
)

//SetupRouting ルーティング
func SetupRouting(e *echo.Echo,client Traq) {
	api := e.Group("/api")
	{
		api.GET("/callback",CallbackHandler)
	}
}