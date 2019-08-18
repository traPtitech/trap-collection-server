package router

import (
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
)

//PostGameHandler gameをアップロードする時のメソッド
func PostGameHandler(c echo.Context) error {
	gameFile, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(game)")
	}

	gameName, err := c.FormFile("name")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(name)")
	}

	filename, err := gameName.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in opening file(name)")
	}

	buf := make([]byte, 1024)
	var n int
	var name string
	for {
		n, err = filename.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		name = string(buf[:n])
	}

	b, err := repository.IsThereGame(name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name exists")
	}

	if b {
		return c.String(http.StatusInternalServerError, "A game with the same name exists")
	}

	src, err := gameFile.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in creating file")
	}
	path := "game/" + name + ".zip"

	_, err = os.Stat(path)
	if err == nil {
		return c.String(http.StatusInternalServerError, "A game with the same name exists")
	}

	dstFile, err := os.Create(path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in creating fileplace")
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, src)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in copying file")
	}

	err = repository.AddGame(name, path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in inserting in db")
	}

	return c.String(http.StatusOK, "game has uploaded")
}

//PutGameHandler gameを更新するときのメソッド
func PutGameHandler(c echo.Context) error {
	gameFile, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(game)")
	}

	gameName, err := c.FormFile("name")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(name)")
	}

	filename, err := gameName.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in opening file(name)")
	}

	buf := make([]byte, 1024)
	var n int
	var name string
	for {
		n, err = filename.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		name = string(buf[:n])
	}

	b, err := repository.IsThereGame(name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name exists")
	}

	if b {
		return c.String(http.StatusInternalServerError, "A game with the same name exists")
	}

	src, err := gameFile.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in creating file")
	}
	path := "game/" + name + ".zip"

	_, err = os.Stat(path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "A game with the same name does not exist")
	}

	dstFile, err := os.Create(path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in creating fileplace")
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, src)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in copying file")
	}

	err = repository.UpdateGame(name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in updating db")
	}

	return c.String(http.StatusOK, "game has updated")
}

//DeleteGameHandler gameを削除するメソッド
func DeleteGameHandler(c echo.Context) error {
	type GameName struct {
		Name string `json:"name,omitempty" db:"name"`
	}
	name := GameName{}
	c.Bind(&name)

	b, err := repository.IsThereGame(name.Name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name exists")
	}

	if b {
		return c.String(http.StatusInternalServerError, "A game with the same name exists")
	}

	path := "game/" + name.Name + ".zip"

	_, err = os.Stat(path)
	if err != nil {
		return c.String(http.StatusInternalServerError, "A game with the same name does not exist")
	}

	if err := os.Remove(path); err != nil {
		panic(err)
	}

	err = repository.DeleteGame(name.Name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting game from db")
	}

	return c.String(http.StatusOK, "game has deleted")
}

//GetGameNameListHandler gameの名前の一覧を取得するメソッド
func GetGameNameListHandler(c echo.Context) error {
	gameNames, err := repository.GetGameNameList()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting the list of game`s name from db")
	}

	return c.JSON(http.StatusOK, gameNames)
}
