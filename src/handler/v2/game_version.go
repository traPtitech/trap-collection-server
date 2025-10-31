package v2

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVersion struct {
	gameVersionService service.GameVersionV2
}

func NewGameVersion(gameVersionService service.GameVersionV2) *GameVersion {
	return &GameVersion{
		gameVersionService: gameVersionService,
	}
}

// ゲームバージョン一覧の取得
// (GET /games/{gameID}/versions)
func (gameVersion *GameVersion) GetGameVersion(c echo.Context, gameID openapi.GameIDInPath, params openapi.GetGameVersionParams) error {
	var limit uint
	if params.Limit != nil {
		limit = uint(*params.Limit)
	}

	var offset uint
	if params.Offset != nil {
		offset = uint(*params.Offset)
	}

	var param *service.GetGameVersionsParams
	if limit != 0 || offset != 0 {
		param = &service.GetGameVersionsParams{
			Limit:  limit,
			Offset: offset,
		}
	}

	num, versions, err := gameVersion.gameVersionService.GetGameVersions(
		c.Request().Context(),
		values.NewGameIDFromUUID(gameID),
		param,
	)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidLimit) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid limit")
	}
	if err != nil {
		log.Printf("error: failed to get game versions: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game versions")
	}

	resVersions := make([]openapi.GameVersion, 0, len(versions))
	for _, version := range versions {
		var resURL *openapi.GameURL
		urlValue, ok := version.Assets.URL.Value()
		if ok {
			v := (*url.URL)(urlValue).String()
			resURL = &v
		}

		var resFiles *openapi.GameVersionFiles
		windows, windowsOk := version.Assets.Windows.Value()
		mac, macOk := version.Assets.Mac.Value()
		jar, jarOk := version.Assets.Jar.Value()
		if windowsOk || macOk || jarOk {
			resFiles = &openapi.GameVersionFiles{}

			if windowsOk {
				v := (uuid.UUID)(windows)
				resFiles.Win32 = &v
			}

			if macOk {
				v := (uuid.UUID)(mac)
				resFiles.Darwin = &v
			}

			if jarOk {
				v := (uuid.UUID)(jar)
				resFiles.Jar = &v
			}
		}

		resVersions = append(resVersions, openapi.GameVersion{
			Id:          openapi.GameVersionID(version.GetID()),
			Name:        string(version.GetName()),
			Description: string(version.GetDescription()),
			CreatedAt:   version.GetCreatedAt(),
			ImageID:     openapi.GameImageID(version.ImageID),
			VideoID:     openapi.GameVideoID(version.VideoID),
			Url:         resURL,
			Files:       resFiles,
		})
	}

	return c.JSON(http.StatusOK, openapi.GetGameVersionsResponse{
		Num:      int(num),
		Versions: resVersions,
	})
}

// ゲームのバージョンの作成
// (POST /games/{gameID}/versions)
func (gameVersion *GameVersion) PostGameVersion(c echo.Context, gameID openapi.GameIDInPath) error {
	var newGameVersion openapi.NewGameVersion
	err := c.Bind(&newGameVersion)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	name := values.NewGameVersionName(newGameVersion.Name)
	err = name.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid name: %s", err.Error()))
	}

	var reqURL types.Option[values.GameURLLink]
	if newGameVersion.Url != nil {
		urlValue, err := url.Parse(*newGameVersion.Url)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid url")
		}

		reqURL = types.NewOption(values.NewGameURLLink(urlValue))
	}

	var (
		reqWindows types.Option[values.GameFileID]
		reqMac     types.Option[values.GameFileID]
		reqJar     types.Option[values.GameFileID]
	)
	if newGameVersion.Files != nil {
		if newGameVersion.Files.Win32 != nil {
			reqWindows = types.NewOption(values.NewGameFileIDFromUUID(*newGameVersion.Files.Win32))
		}

		if newGameVersion.Files.Darwin != nil {
			reqMac = types.NewOption(values.NewGameFileIDFromUUID(*newGameVersion.Files.Darwin))
		}

		if newGameVersion.Files.Jar != nil {
			reqJar = types.NewOption(values.NewGameFileIDFromUUID(*newGameVersion.Files.Jar))
		}
	}

	gameVersionInfo, err := gameVersion.gameVersionService.CreateGameVersion(
		c.Request().Context(),
		values.NewGameIDFromUUID(gameID),
		name,
		values.NewGameVersionDescription(newGameVersion.Description),
		values.GameImageIDFromUUID(newGameVersion.ImageID),
		values.NewGameVideoIDFromUUID(newGameVersion.VideoID),
		&service.Assets{
			URL:     reqURL,
			Windows: reqWindows,
			Mac:     reqMac,
			Jar:     reqJar,
		},
	)
	switch {
	case errors.Is(err, service.ErrInvalidGameID):
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	case errors.Is(err, service.ErrInvalidGameImageID):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid imageID")
	case errors.Is(err, service.ErrInvalidGameVideoID):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid videoID")
	case errors.Is(err, service.ErrInvalidGameFileID):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid fileID")
	case errors.Is(err, service.ErrInvalidGameFileType):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid fileType")
	case errors.Is(err, service.ErrNoAsset):
		return echo.NewHTTPError(http.StatusBadRequest, "no assets")
	case errors.Is(err, service.ErrDuplicateGameVersion):
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate game version")
	case err != nil:
		log.Printf("failed to create game version: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create game version")
	}

	var resURL *openapi.GameURL
	urlValue, ok := gameVersionInfo.Assets.URL.Value()
	if ok {
		v := (*url.URL)(urlValue).String()
		resURL = &v
	}

	var resFiles *openapi.GameVersionFiles
	windows, windowsOk := gameVersionInfo.Assets.Windows.Value()
	mac, macOk := gameVersionInfo.Assets.Mac.Value()
	jar, jarOk := gameVersionInfo.Assets.Jar.Value()
	if windowsOk || macOk || jarOk {
		resFiles = &openapi.GameVersionFiles{}

		if windowsOk {
			v := (uuid.UUID)(windows)
			resFiles.Win32 = &v
		}

		if macOk {
			v := (uuid.UUID)(mac)
			resFiles.Darwin = &v
		}

		if jarOk {
			v := (uuid.UUID)(jar)
			resFiles.Jar = &v
		}
	}

	return c.JSON(http.StatusCreated, openapi.GameVersion{
		Id:          openapi.GameVersionID(gameVersionInfo.GameVersion.GetID()),
		Name:        string(gameVersionInfo.GameVersion.GetName()),
		Description: string(gameVersionInfo.GameVersion.GetDescription()),
		CreatedAt:   gameVersionInfo.GetCreatedAt(),
		ImageID:     openapi.GameImageID(gameVersionInfo.ImageID),
		VideoID:     openapi.GameVideoID(gameVersionInfo.VideoID),
		Url:         resURL,
		Files:       resFiles,
	})
}

// ゲームの最新バージョンの取得
// (GET /games/{gameID}/versions/latest)
func (gameVersion *GameVersion) GetLatestGameVersion(ctx echo.Context, gameID openapi.GameIDInPath) error {
	gameVersionInfo, err := gameVersion.gameVersionService.GetLatestGameVersion(
		ctx.Request().Context(),
		values.NewGameIDFromUUID(gameID),
	)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrNoGameVersion) {
		return echo.NewHTTPError(http.StatusNotFound, "no game version")
	}
	if err != nil {
		log.Printf("failed to get latest game version: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get latest game version")
	}

	var resURL *openapi.GameURL
	urlValue, ok := gameVersionInfo.Assets.URL.Value()
	if ok {
		v := (*url.URL)(urlValue).String()
		resURL = &v
	}

	var resFiles *openapi.GameVersionFiles
	windows, windowsOk := gameVersionInfo.Assets.Windows.Value()
	mac, macOk := gameVersionInfo.Assets.Mac.Value()
	jar, jarOk := gameVersionInfo.Assets.Jar.Value()
	if windowsOk || macOk || jarOk {
		resFiles = &openapi.GameVersionFiles{}

		if windowsOk {
			v := (uuid.UUID)(windows)
			resFiles.Win32 = &v
		}

		if macOk {
			v := (uuid.UUID)(mac)
			resFiles.Darwin = &v
		}

		if jarOk {
			v := (uuid.UUID)(jar)
			resFiles.Jar = &v
		}
	}

	return ctx.JSON(http.StatusOK, openapi.GameVersion{
		Id:          openapi.GameVersionID(gameVersionInfo.GameVersion.GetID()),
		Name:        string(gameVersionInfo.GameVersion.GetName()),
		Description: string(gameVersionInfo.GameVersion.GetDescription()),
		CreatedAt:   gameVersionInfo.GetCreatedAt(),
		ImageID:     openapi.GameImageID(gameVersionInfo.ImageID),
		VideoID:     openapi.GameVideoID(gameVersionInfo.VideoID),
		Url:         resURL,
		Files:       resFiles,
	})
}
