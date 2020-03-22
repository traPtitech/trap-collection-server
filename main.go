package main

import (
	"os"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/srinathgs/mysqlstore"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/router"
)

func main() {
	err := model.EstablishConoHa()
	if err != nil {
		panic(err)
	}
	db, err := model.EstablishDB()
	if err != nil {
		panic(err)
	}
	store, err := mysqlstore.NewMySQLStoreFromConnection(db, "sessions", "/", 60*60*24*14, []byte("secret-token"))
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))

	router.InitRouter()

	if os.Getenv("COLLECTION_ENV") == "test" {
		mockClient := &router.MockTraqClient{
			User: router.User{
				ID:   os.Getenv("USER_ID"),
				Name: os.Getenv("USER_Name"),
			},
		}
		router.SetupRouting(e, mockClient)
	} else {
		router.SetupRouting(e, &router.TraqClient{})
	}

	e.Start(os.Getenv("PORT"))
}
