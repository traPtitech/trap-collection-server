package router

import (
	echo "github.com/labstack/echo/v4"
)

//SetupRouting ルーティング
func SetupRouting(e *echo.Echo, client Traq) {
	apiLancherAuth := e.Group("/api", MiddlewareAuthLancher)
	{
		apiVersion := apiLancherAuth.Group("/versions")
		{
			apiVersion.GET("/check", GetCheckListHandler)
		}
	}
}
