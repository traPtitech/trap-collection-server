package router

import (
	"crypto/md5"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
)

//CheckHandler gameに破損・更新がないか確認するメソッド
func CheckHandler(c echo.Context) error {
	games, err := repository.GameCheckList()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting the list of game from db")
	}

	type check struct {
		GameID string `json:"gameId,omitempty" db:"game_id"`
		Name   string `json:"name,omitempty" db:"name"`
		Md5    string `json:"md5"`
	}

	checks := []check{}
	for i, v := range games {
		checks[i].GameID = v.GameID
		checks[i].Name = v.Name
		file, err := os.Open(v.Path)
		if err != nil {
			return c.String(http.StatusInternalServerError, "something wrong in opening file")
		}

		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n == 0 {
				break
			}
			if err != nil {
				break
			}
			md5:=md5.Sum(buf)
			checks[i].Md5 = string(md5[:n])
		}
	}

	type checkList struct {
		List      []check   `json:"list,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty"`
	}
	checkLists := checkList{}
	checkLists.List = checks
	updatedAt, err := repository.LastUpdatedAt()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting last updated time from db")
	}
	checkLists.UpdatedAt = updatedAt

	return c.JSON(http.StatusOK, checkLists)
}

//DownloadHandler ダウンロードのメソッド
func DownloadHandler(c echo.Context) error {
	gameName := c.Param("name")
	path, err := repository.GetPath(gameName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting path from db")
	}

	return c.File(path)
}
