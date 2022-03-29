package main

import (
	"github.com/traPtitech/trap-collection-server/src"
)

func main() {
	service, err := src.InjectAPI()
	if err != nil {
		panic(err)
	}
	defer service.DB.Close()

	err = service.Start()
	if err != nil {
		panic(err)
	}
}
