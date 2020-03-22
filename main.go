package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/router"
)

func main() {
	err := model.EstablishConoHa()
	if err != nil {
		panic(err)
	}
	err = model.EstablishDB()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	router.InitRouter()

	if os.Getenv("COLLECTION_ENV") == "test" {
		mockClient := &router.MockTraqClient{
			User: router.User{
				ID: os.Getenv("USER_ID"),
				Name: os.Getenv("USER_Name"),
			},
		}
		router.SetupRouting(e, mockClient)
	} else {
		router.SetupRouting(e, &router.TraqClient{})
	}

	e.Start(os.Getenv("PORT"))
}