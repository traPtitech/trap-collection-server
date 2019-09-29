package router

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/repository"
)

//CheckHandler gameに破損・更新がないか確認するメソッド
func CheckHandler(c echo.Context) error {
	version := c.Param("version")
	checks, err := repository.GameCheckListByVersion(version)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting the list of game from db")
	}

	type check struct {
		GameID string `json:"id,omitempty" db:"id"`
		Name   string `json:"name,omitempty" db:"name"`
		Md5    string `json:"md5" db:"md5"`
	}

	type checkList struct {
		List      []model.GameCheck `json:"list,omitempty"`
		UpdatedAt time.Time         `json:"updatedAt,omitempty"`
	}
	checkLists := checkList{}
	checkLists.List = checks
	updatedAt, err := repository.GetLastUpdatedAtByVersion(version)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting last updated time from db")
	}
	checkLists.UpdatedAt = updatedAt

	return c.JSON(http.StatusOK, checkLists)
}

//DownloadHandler ダウンロードのメソッド
func DownloadHandler(c echo.Context) error {
	gameID := c.Param("id")
	gameName, err := repository.GameIDToName(gameID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting the game name")
	}
	game, err := repository.DownloadGame(gameName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting the game file")
	}

	return c.Blob(200, "application/zip", game)
}
