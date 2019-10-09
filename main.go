package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/traPtitech/trap-collection-server/repository"
	"github.com/traPtitech/trap-collection-server/router"
)

func main() {
	//err := repository.EstablishConoHa()
	//if err != nil {
	//	panic(err)
	//}
	err := repository.EstablishDB()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	router.SetRouting(e)

	e.Start(os.Getenv("PORT"))
}
