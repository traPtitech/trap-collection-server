package router

import (
	"fmt"
	"strconv"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Version versionの構造体
type Version struct {
	LauncherAuthBase
	openapi.VersionApi
}

// GetVersion GET /version/:launcherVersionIDの処理部分
func (*Version) GetVersion(strLauncherVersion string) (openapi.VersionDetails, sessionMap, error) {
	launcherVersionID, err := strconv.Atoi(strLauncherVersion)
	if err != nil {
		return openapi.VersionDetails{}, sessionMap{}, fmt.Errorf("Failed In Comverting Launcher Version ID:%w", err)
	}
	launcherVersion, err := model.GetLauncherVersionDetailsByID(uint(launcherVersionID))
	if err != nil {
		return openapi.VersionDetails{}, sessionMap{}, fmt.Errorf("Failed In Getting Launcher Version ID:%w", err)
	}

	return launcherVersion, sessionMap{}, nil
}

// GetCheckList GET /versions/checkの処理部分
func (v *Version) GetCheckList(operationgSystem string, sess sessionMap) ([]openapi.CheckItem, sessionMap, error) {
	versionID, err := v.getVersionID(sess)
	if err != nil {
		return []openapi.CheckItem{}, sessionMap{}, fmt.Errorf("Failed In Getting VersionID: %w", err)
	}
	checkList, err := model.GetCheckList(versionID, operationgSystem)
	if err != nil {
		return []openapi.CheckItem{}, sessionMap{}, fmt.Errorf("Failed In Getting CheckList: %w", err)
	}
	return checkList, sessionMap{}, nil
}

// GetQuestions GET /versions/questionの処理部分
func (v *Version) GetQuestions(sess sessionMap) ([]openapi.Question, sessionMap, error) {
	versionID, err := v.getVersionID(sess)
	if err != nil {
		return []openapi.Question{}, sessionMap{}, fmt.Errorf("Failed In Getting VersionID: %w", err)
	}
	questions, err := model.GetQuestions(versionID)
	if err != nil {
		return []openapi.Question{}, sessionMap{}, fmt.Errorf("Failed In Getting Questions: %w", err)
	}
	return questions, sessionMap{}, nil
}
