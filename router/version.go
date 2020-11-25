package router

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
)

// Version versionの構造体
type Version struct {
	db           model.DBMeta
	launcherAuth base.LauncherAuth
	openapi.VersionApi
}

func newVersion(db model.DBMeta, launcherAuth base.LauncherAuth) openapi.VersionApi {
	version := new(Version)

	version.db = db
	version.launcherAuth = launcherAuth

	return version
}

// GetVersions GET /versionsの処理部分
func (v *Version) GetVersions() ([]*openapi.Version, error) {
	versions, err := v.db.GetLauncherVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher versions: %w", err)
	}

	return versions, nil
}

// PostVersion POST /versions
func (v *Version) PostVersion(newVersion *openapi.NewVersion) (*openapi.VersionMeta, error) {
	version, err := v.db.InsertLauncherVersion(newVersion.Name, newVersion.AnkeTo)
	if err != nil {
		return nil, fmt.Errorf("failed to insert a lancher version: %w", err)
	}

	return version, nil
}

// GetVersion GET /versions/:launcherVersionIDの処理部分
func (v *Version) GetVersion(strLauncherVersion string) (*openapi.VersionDetails, error) {
	launcherVersionID, err := strconv.Atoi(strLauncherVersion)
	if err != nil {
		return &openapi.VersionDetails{}, fmt.Errorf("Failed In Comverting Launcher Version ID:%w", err)
	}

	launcherVersion, err := v.db.GetLauncherVersionDetailsByID(uint(launcherVersionID))
	if err != nil {
		return &openapi.VersionDetails{}, fmt.Errorf("Failed In Getting Launcher Version ID:%w", err)
	}

	return launcherVersion, nil
}

// PostGameToVersion POST /version/:launcherVersionID/gameの処理部分
func (v *Version) PostGameToVersion(launcherVersionID string, gameIDs *openapi.GameIDs) (*openapi.Version, error) {
	intLauncherVersionID, err := strconv.Atoi(launcherVersionID)
	if err != nil {
		return nil, errors.New("invalid launcherVersionID")
	}

	err = v.db.CheckGameIDs(gameIDs.GameIDs)
	if err != nil {
		gameIDErr := &model.GameIDsError{}
		if errors.As(err, &gameIDErr) {
			return nil, gameIDErr
		}

		return nil, fmt.Errorf("failed to check gameIDs: %w", err)
	}

	version, err := v.db.InsertGamesToLauncherVersion(intLauncherVersionID, gameIDs.GameIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to insert games to version: %w", err)
	}

	return version, nil
}

// GetCheckList GET /versions/checkの処理部分
func (v *Version) GetCheckList(operationgSystem string, c echo.Context) ([]*openapi.CheckItem, error) {
	versionID, err := v.launcherAuth.GetVersionID(c)
	if err != nil {
		return []*openapi.CheckItem{}, fmt.Errorf("Failed In Getting VersionID: %w", err)
	}

	checkList, err := v.db.GetCheckList(versionID, operationgSystem)
	if err != nil {
		return []*openapi.CheckItem{}, fmt.Errorf("Failed In Getting CheckList: %w", err)
	}

	return checkList, nil
}
