package router

import (
	"fmt"
	"strconv"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Version versionの構造体
type Version struct {
	openapi.VersionApi
}

// GetVersion GET /version/{launcherVersionID}のハンドラー
func (v Version) GetVersion(strLauncherVersion string) (openapi.VersionDetails, map[interface{}]interface{},error) {
	launcherVersionID,err := strconv.Atoi(strLauncherVersion)
	if err != nil {
		return openapi.VersionDetails{}, map[interface{}]interface{}{}, fmt.Errorf("Failed In Comverting Launcher Version ID:%w",err)
	}
	launcherVersion,err := model.GetLauncherVersionDetailsByID(uint(launcherVersionID))
	if err != nil {
		return openapi.VersionDetails{}, map[interface{}]interface{}{},fmt.Errorf("Failed In Getting Launcher Version ID:%w",err)
	}

	return launcherVersion, map[interface{}]interface{}{}, nil
}
