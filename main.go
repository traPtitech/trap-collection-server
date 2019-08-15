package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/traPtitech/trap-collection-server/router"
)

func main() {
	store, err := model.Establish()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))

	e.POST("/game", router.PostGameHandler)
	e.PUT("/game", router.PutGameHandler)
	e.DELETE("/game", router.DeleteGameHandler)
	e.GET("/game", GetGameListHandler)
	e.POST("/check", CheckHandler)
	e.GET("/download/:name", DownloadHandler)

	e.Start(":11400")
}
