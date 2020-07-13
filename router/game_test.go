package router

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
	"github.com/traPtitech/trap-collection-server/storage"
)

func TestGame(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := model.NewDBMock(ctrl)
	str:= storage.NewMockStorage(ctrl)

	oauth := base.NewMockOAuth(ctrl)

	gameID := "72c0c88c-27fd-4b58-b08e-e3307d2c17df"
	createdAt, err := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Feb 3, 2013 at 7:54pm (PST)")
	if err != nil {
		t.Fatalf("Failed In Parse Time: %#v", err)
	}
	openapiGame := &openapi.Game{
		Id: "72c0c88c-27fd-4b58-b08e-e3307d2c17df",
		Name: "ClayPlatesStory",
		CreatedAt: createdAt,
		Version: &openapi.GameVersion{
			Id: 1,
			Name: "v1.0",
			Description: "ClayPlatesStory",
			CreatedAt: createdAt,
		},
	}

	game := newGame(db, oauth, str)
	db.MockGameMeta.
		EXPECT().
		GetGameInfo(gameID).
		Return(openapiGame, nil)

	_, err = game.GetGame(gameID)
	if err != nil {
		t.Fatalf("Unexpected GetGame Error: %#v", err)
	}

	oss := []string{"windows", "mac"}
	fileTypes := []string{"jar", "mac"}
	exts := []string{"jar", "zip"}
	for i, fileType := range fileTypes {
		db.MockGameVersionMeta.
			EXPECT().
			GetGameType(gameID, oss[i]).
			Return(fileType, nil)
	}

	gameFileNames := make([]string, 0, len(oss))
	for i, os := range oss {
		res, err := game.getGameFileName(gameID, os)
		if err != nil {
			t.Fatalf("Unexpected getGameFileName Error: %#v", err)
		}

		expect := gameID + "_game." + exts[i]
		if res != expect {
			t.Fatalf("Unexpedcted gameFileName %s, Expected %s", res, expect)
		}

		gameFileNames = append(gameFileNames, res)
	}

	expectStr := "test"
	db.MockGameVersionMeta.
		EXPECT().
		GetGameType(gameID, oss[0]).
		Return(fileTypes[0], nil)
	str.
		EXPECT().
		Open(gameFileNames[0]).
		Return(ioutil.NopCloser(strings.NewReader(expectStr)), nil)
	
	res, err := game.GetGameFile(gameID, oss[0])
	if err != nil {
		t.Fatalf("Unexpected GetGameFile Error: %#v", err)
	}
	buf := new(bytes.Buffer)
	_,err = buf.ReadFrom(res)
	if err != nil {
		t.Fatalf("Unexpected File Reed Error: %#v", err)
	}
	strRes := buf.String()
	if strRes != expectStr {
		t.Fatalf("Unexpected File Value %s, Expected %s", strRes, expectStr)
	}
}
