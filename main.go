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
	env := os.Getenv("COLLECTION_ENV")
	err := model.EstablishConoHa()
	if err != nil {
		panic(err)
	}

	db, err := model.EstablishDB(false)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db2,err := model.EstablishDB(true)
	if err != nil {
		panic(err)
	}
	defer db2.Close()

	if env == "development" {
		db.LogMode(true)
	}

	store, err := mysqlstore.NewMySQLStoreFromConnection(db.DB(), "sessions", "/", 60*60*24*14, []byte("secret-token"))
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(session.Middleware(store))

	if env == "test" {
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

	err = router.InitRouter()
	if err != nil {
		panic(err)
	}

	e.Start(os.Getenv("PORT"))
}
