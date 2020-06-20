package router

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/storage"
)

func TestGame(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := model.NewDBMock(ctrl)

	str, err := storage.NewLocalStorage("../upload")
	if err != nil {
		t.Fatalf("Failed In Storage Constructor: %#v", err)
	}

	game := newGame(db, str)

	_,err = game.GetGameFile("test","win")
	if err != nil {
		t.Fatalf("Failed In Getting Game File: %#v", err)
	}
}
