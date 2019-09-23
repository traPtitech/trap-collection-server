package router

import (
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
	"github.com/traPtitech/trap-collection-server/model"
)
//Version バージョンの構造体
type Version struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	QuestionnairID int `json:"questionnair_id,omitempty"`
	StartPeriod time.Time `json:"start_period,omitempty"`
	EndPeriod time.Time `json:"end_period,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	SpecialList []string `json:"special_list,omitempty"`
}

//PostVersionHandler バージョンの追加
func PostVersionHandler(c echo.Context) error {
	version := Version{}
	err := c.Bind(&version)
	if err!= nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}

	b,err := repository.IsThereVersion(version.Name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if is there the version")
	}
	if b {
		return c.String(http.StatusNotAcceptable, "same name of version exists")
	}

	if version.Type=="for sale" {
		err = repository.InsertVersionForSale(version.Name,version.StartPeriod,version.EndPeriod,version.StartTime)
	}else if version.Type=="not for sale" {
		err = repository.InsertVersionNotForSale(version.Name, version.QuestionnairID, version.StartPeriod,version.EndPeriod,version.StartTime)
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in inserting version")
	}

	return nil
}

//PutVersionHandler バージョンの追加
func PutVersionHandler(c echo.Context) error {
	version := Version{}
	err := c.Bind(&version)
	if err!= nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}


	b,err := repository.IsThereVersion(version.Name)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if is there the version")
	}
	if !b {
		return c.String(http.StatusNotAcceptable, "same name of version does not exist")
	}

	if version.Type=="for sale" {
		err = repository.UpdateVersionForSale(version.Name,version.StartPeriod,version.EndPeriod,version.StartTime)
	}else if version.Type=="not for sale" {
		err = repository.UpdateVersionNotForSale(version.Name, version.QuestionnairID, version.StartPeriod,version.EndPeriod,version.StartTime)
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in inserting version")
	}
	err = repository.DeleteSpecialByPeriod(version.Name,version.StartPeriod,version.EndPeriod)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting special case")
	}

	return nil
}

//DeleteVersionHandler バージョンの削除
func DeleteVersionHandler(c echo.Context) error {
	type VersionName struct {
		Type string `json:"type,omitempty"`
		Name string `json:"name,omitempty"`
	}
	version := VersionName{}
	err := c.Bind(&version)
	if err!= nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}

	if version.Type=="for sale" {
		err = repository.DeleteVersionForSale(version.Name)
	} else if version.Type=="not for sale" {
		err = repository.DeleteVersionNotForSale(version.Name)
	}
	if err!=nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting version")
	}

	err = repository.DeleteSpecialByVersion(version.Name)
	if err!=nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting special cases")
	}

	return nil
}

//GetVersionForSaleListHandler 販売用バージョン一覧の取得
func GetVersionForSaleListHandler(c echo.Context) error {
	type VersionList struct {
		List []model.VersionForSale `json:"list,omitempty"`
	}
	versions := []model.VersionForSale{}
	versions,err := repository.VersionForSaleList()
	if err!=nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting versions for sale")
	}

	versionList := VersionList{}
	versionList.List = versions
	return c.JSON(http.StatusOK, &versionList)
}

//GetVersionNotForSaleListHandler 工大祭用バージョン一覧の取得
func GetVersionNotForSaleListHandler(c echo.Context) error {
	type VersionList struct {
		List []model.VersionNotForSale `json:"list,omitempty"`
	}
	versions := []model.VersionNotForSale{}
	versions,err := repository.VersionNotForSaleList()
	if err!=nil {
		return c.String(http.StatusInternalServerError, "something wrong in getting versions for sale")
	}

	versionList := VersionList{}
	versionList.List = versions
	return c.JSON(http.StatusOK, &versionList)
}