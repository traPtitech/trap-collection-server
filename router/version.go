package router

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/repository"
)

//Version バージョンの構造体
type Version struct {
	Type           string          `json:"type,omitempty"`
	Name           string          `json:"name,omitempty"`
	QuestionnairID int             `json:"questionnair_id,omitempty"`
	StartPeriod    time.Time       `json:"start_period,omitempty"`
	EndPeriod      time.Time       `json:"end_period,omitempty"`
	StartTime      time.Time       `json:"start_time,omitempty"`
	SpecialList    []model.Special `json:"special_list,omitempty"`
}

//PostVersionHandler バージョンの追加
func PostVersionHandler(c echo.Context) error {
	version := Version{}
	err := c.Bind(&version)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}

	b, err := repository.IsThereVersion(version.Name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if is there the version")
	}
	if b {
		return c.String(http.StatusNotAcceptable, "same name of version exists")
	}

	var id string
	if version.Type == "for sale" {
		id, err = repository.InsertVersionForSale(version.Name, version.StartPeriod, version.EndPeriod, version.StartTime)
	} else if version.Type == "not for sale" {
		id, err = repository.InsertVersionNotForSale(version.Name, version.QuestionnairID, version.StartPeriod, version.EndPeriod, version.StartTime)
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in inserting version")
	}

	//明らかなN+1、時間ができたら直す、たぶんtransactionとかを使った方がいい
	for _, v := range version.SpecialList {
		b, err := repository.IsThereGame(v.GameName)
		if err != nil {
			return c.String(http.StatusInternalServerError, "something wrong in checking if is there the game")
		}
		if !b {
			return c.String(http.StatusInternalServerError, "the game does not exist")
		}

		b, err = repository.IsThereSpecial(id, v.GameName)
		if err != nil {
			return c.String(http.StatusInternalServerError, "something wrong in checking if is there the special case")
		}
		if !b {
			continue
		}

		err = repository.InsertSpecial(id, v.GameName, v.InOut)
		if err != nil {
			return c.String(http.StatusInternalServerError, "something wrong in inserting special case")
		}
	}

	return c.String(http.StatusOK, "version has created")
}

//PutVersionHandler バージョンの修正
func PutVersionHandler(c echo.Context) error {
	version := Version{}
	err := c.Bind(&version)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}
	id := c.Param("id")

	b, err := repository.IsThereVersion(version.Name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if is there the version")
	}
	if !b {
		return c.String(http.StatusNotAcceptable, "same name of version does not exist")
	}

	if version.Type == "for sale" {
		err = repository.UpdateVersionForSale(id, version.Name, version.StartPeriod, version.EndPeriod, version.StartTime)
	} else if version.Type == "not for sale" {
		err = repository.UpdateVersionNotForSale(id, version.Name, version.QuestionnairID, version.StartPeriod, version.EndPeriod, version.StartTime)
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in inserting version")
	}
	err = repository.DeleteSpecialByPeriod(version.Name, version.StartPeriod, version.EndPeriod)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting special case")
	}

	return c.String(http.StatusOK, "version has updated")
}

//DeleteVersionHandler バージョンの削除
func DeleteVersionHandler(c echo.Context) error {
	id := c.Param("id")

	var err error
	err = repository.DeleteVersionForSale(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting version")
	}
	err = repository.DeleteVersionNotForSale(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting version")
	}

	err = repository.DeleteSpecialByVersion(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting special cases")
	}

	return c.String(http.StatusOK, "version has deleted")
}

//GetVersionForSaleListHandler 販売用バージョン一覧の取得
func GetVersionForSaleListHandler(c echo.Context) error {
	versions := []model.VersionForSale{}
	versions, err := repository.VersionForSaleList()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting versions for sale")
	}
	return c.JSON(http.StatusOK, versions)
}

//GetVersionNotForSaleListHandler 工大祭用バージョン一覧の取得
func GetVersionNotForSaleListHandler(c echo.Context) error {
	versions := []model.VersionNotForSale{}
	versions, err := repository.VersionNotForSaleList()
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting versions for sale")
	}
	return c.JSON(http.StatusOK, versions)
}

//GetGameListHandler ゲーム一覧の取得
func GetGameListHandler(c echo.Context) error {
	id := c.Param("id")
	games, err := repository.GetGameListByVersion(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting games")
	}
	return c.JSON(http.StatusOK, games)
}

//GetNonGameListHandler バージョン外ゲーム一覧の取得
func GetNonGameListHandler(c echo.Context) error {
	id := c.Param("id")
	games, err := repository.GetNonGameListByVersion(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting games")
	}
	return c.JSON(http.StatusOK, games)
}

//GetQuestionnaireHandler アンケートの取得
func GetQuestionnaireHandler(c echo.Context) error {
	type Questionnaire struct {
		ID int `json:"questionnaire,omitempty"`
	}
	id := c.Param("id")
	questionnairID, err := repository.GetQuestionnaireByVersionNotForSale(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting questionnaire")
	}
	questionnaire := Questionnaire{}
	questionnaire.ID = questionnairID
	return c.JSON(http.StatusOK, questionnaire)
}
