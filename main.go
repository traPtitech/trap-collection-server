package main

import (
	"github.com/traPtitech/trap-collection-server/src/wire"
)

func main() {
	app, err := wire.InjectApp()
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}
}
