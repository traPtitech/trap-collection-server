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
	games, err := repository.GetGameList()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting the list of game from db")
	}

	type check struct {
		GameID string `json:"gameId,omitempty" db:"game_id"`
		Name   string `json:"name,omitempty" db:"name"`
		Md5    string `json:"md5"`
	}

	checks := []check{}
	for i, v = range games {
		checks[i].GameID = v.GameID
		checks[i].Name = v.Name
		file, err := os.Open(v.Path)

		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n == 0 {
				break
			}
			if err != nil {
				break
			}
			checks[i].Md5 = md5.Sum(buf)
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
	checkList.UpdatedAt = updatedAt

	return c.JSON(http.StatusOK, checkLists)
}

//DownloadHandler ダウンロードのメソッド
func DownloadHandler(c echo.Context) error {
	gameName := c.Param("name")
	path, err = repository.GetPath(gameName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting path from db")
	}

	return c.File(path)
}
