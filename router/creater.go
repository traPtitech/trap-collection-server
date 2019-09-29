package router

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"unsafe"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
)

//PostGameHandler gameをアップロードする時のメソッド
func PostGameHandler(c echo.Context) error {
	containerName := "game0"
	gameFile, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(game)")
	}

	file, err := gameFile.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in opening file")
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading file")
	}
	md5 := md5.Sum(buf)

	gameName, err := c.FormFile("name")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(name)")
	}

	filename, err := gameName.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in opening file(name)")
	}

	buf, err = ioutil.ReadAll(filename)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading file(name)")
	}
	name := *(*string)(unsafe.Pointer(&buf))

	b := repository.IsThereGame(name)
	if b {
		return c.String(http.StatusInternalServerError, "A game with the same name exists")
	}

	nameOfFile := name + ".zip"
	b, err = repository.IsThereFile(nameOfFile)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name of file exists")
	}

	if b {
		return c.String(http.StatusInternalServerError, "A game with the same name of file exists")
	}

	err = repository.UploadGame(gameFile, nameOfFile, containerName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in uploading file")
	}

	err = repository.AddGame(name, containerName, nameOfFile, md5)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in inserting in db")
	}

	return c.String(http.StatusOK, "game has uploaded")
}

//PutGameHandler gameを更新するときのメソッド
func PutGameHandler(c echo.Context) error {
	containerName := "game0"
	gameFile, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(game)")
	}

	file, err := gameFile.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in opening file")
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading file")
	}
	md5 := md5.Sum(buf)

	gameName, err := c.FormFile("name")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(name)")
	}

	filename, err := gameName.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in opening file(name)")
	}

	buf, err = ioutil.ReadAll(filename)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading file(name)")
	}
	name := *(*string)(unsafe.Pointer(&buf))

	b := repository.IsThereGame(name)
	if !b {
		return c.String(http.StatusInternalServerError, "A game with the same name does not exist")
	}

	nameOfFile := name + ".zip"
	b, err = repository.IsThereFile(nameOfFile)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name of file exists")
	}
	if !b {
		return c.String(http.StatusInternalServerError, "A game with the same name of file does not exist")
	}

	err = repository.UploadGame(gameFile, nameOfFile, containerName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in uploading file")
	}

	err = repository.UpdateGame(name, md5)
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusInternalServerError, "something wrong in updating db")
	}

	return c.String(http.StatusOK, "game has updated")
}

//DeleteGameHandler gameを削除するメソッド(無駄がありまくりなので時間ができたら修正)
func DeleteGameHandler(c echo.Context) error {
	id := c.Param("id")

	name, err := repository.GameIDToName(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting game`s name")
	}

	b := repository.IsThereGame(name)
	if !b {
		return c.String(http.StatusInternalServerError, "A game with the name doesn`t exist")
	}

	nameOfFile := name + ".zip"
	b, err = repository.IsThereFile(nameOfFile)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name of file exists")
	}

	if !b {
		return c.String(http.StatusInternalServerError, "A game with the name of file doesn`t exist")
	}

	err = repository.DeleteGameFromConoHa(name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting file")
	}

	err = repository.DeleteGame(name)
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
