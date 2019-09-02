package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/traPtitech/trap-collection-server/repository"
	"github.com/traPtitech/trap-collection-server/router"
)

func main() {
	err := repository.Establish()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.POST("/game", router.PostGameHandler)
	e.PUT("/game", router.PutGameHandler)
	e.DELETE("/game", router.DeleteGameHandler)
	e.GET("/game", router.GetGameNameListHandler)
	e.POST("/check", router.CheckHandler)
	e.GET("/download/:name", router.DownloadHandler)

	router.SetRouting(e)

	e.Start(os.Getenv("PORT"))
}
