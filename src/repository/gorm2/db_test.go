package gorm2

import (
	"os"
	"testing"

	v1 "github.com/traPtitech/trap-collection-server/src/config/v1"
)

var testDB *DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = NewDB(v1.NewApp(), v1.NewRepositoryGorm2())
	if err != nil {
		panic(err)
	}

	code := m.Run()

	os.Exit(code)
}
