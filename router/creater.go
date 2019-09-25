package router

import (
	"crypto/md5"
	"net/http"

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

	buf := make([]byte, 1024)
	var strmd5 string
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		md5 := md5.Sum(buf)
		strmd5 = string(md5[:n])
	}

	gameName, err := c.FormFile("name")
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in reading fileheader(name)")
	}

	filename, err := gameName.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in opening file(name)")
	}

	buf = make([]byte, 1024)
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

	err = repository.AddGame(name, containerName, nameOfFile, strmd5)
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

	nameOfFile := name + ".zip"
	b, err = repository.IsThereFile(nameOfFile)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name of file exists")
	}

	if b {
		return c.String(http.StatusInternalServerError, "A game with the same name of file exists")
	}

	containerName, err := repository.GetContainerByName(name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting the name of container")
	}
	err = repository.UploadGame(gameFile, nameOfFile, containerName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in uploading file")
	}

	err = repository.UpdateGame(name)
	if err != nil {
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

	b, err := repository.IsThereGame(name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if A game with the same name exists")
	}

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
