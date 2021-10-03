package gorm2

import (
	"os"
	"testing"
)

var testDB *DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = NewDB(true)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	os.Exit(code)
}
