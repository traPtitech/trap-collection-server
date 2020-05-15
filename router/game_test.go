package router

import (
	"testing"

	"github.com/traPtitech/trap-collection-server/storage"
)

func TestGame(t *testing.T) {
	str, err := storage.NewLocalStorage("./upload")
	if err != nil {
		t.Fatalf("Failed In Storage Constructor: %#v", err)
	}
	game := NewGame(&str)

	_,_,err = game.GetGameFile("test","win")
	if err != nil {
		t.Fatalf("Failed In Getting Game File: %#v", err)
	}
}
