package router

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
)

// Version versionの構造体
type Version struct {
	base.LauncherAuth
	openapi.VersionApi
}

// GetVersion GET /version/:launcherVersionIDの処理部分
func (*Version) GetVersion(strLauncherVersion string) (*openapi.VersionDetails, error) {
	launcherVersionID, err := strconv.Atoi(strLauncherVersion)
	if err != nil {
		return &openapi.VersionDetails{}, fmt.Errorf("Failed In Comverting Launcher Version ID:%w", err)
	}
	launcherVersion, err := model.GetLauncherVersionDetailsByID(uint(launcherVersionID))
	if err != nil {
		return &openapi.VersionDetails{}, fmt.Errorf("Failed In Getting Launcher Version ID:%w", err)
	}

	return launcherVersion, nil
}

// GetCheckList GET /versions/checkの処理部分
func (v *Version) GetCheckList(c echo.Context, operationgSystem string) ([]*openapi.CheckItem, error) {
	versionID, err := v.GetVersionID(c)
	if err != nil {
		return []*openapi.CheckItem{}, fmt.Errorf("Failed In Getting VersionID: %w", err)
	}
	checkList, err := model.GetCheckList(versionID, operationgSystem)
	if err != nil {
		return []*openapi.CheckItem{}, fmt.Errorf("Failed In Getting CheckList: %w", err)
	}
	return checkList, nil
}

// GetQuestions GET /versions/questionの処理部分
func (v *Version) GetQuestions(c echo.Context) ([]*openapi.Question, error) {
	versionID, err := v.GetVersionID(c)
	if err != nil {
		return []*openapi.Question{}, fmt.Errorf("Failed In Getting VersionID: %w", err)
	}
	questions, err := model.GetQuestions(versionID)
	if err != nil {
		return []*openapi.Question{}, fmt.Errorf("Failed In Getting Questions: %w", err)
	}
	return questions, nil
}
