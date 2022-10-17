package v2

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVersion struct {
	gameVersionService service.GameVersionV2
	gameVersionUnimplemented
}

func NewGameVersion(gameVersionService service.GameVersionV2) *GameVersion {
	return &GameVersion{
		gameVersionService: gameVersionService,
	}
}

// gameVersionUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameVersionUnimplemented interface {
	// ゲームのバージョンの作成
	// (POST /games/{gameID}/versions)
	PostGameVersion(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームの最新バージョンの取得
	// (GET /games/{gameID}/versions/latest)
	GetLatestGameVersion(ctx echo.Context, gameID openapi.GameIDInPath) error
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

	num, versions, err := gameVersion.gameVersionService.GetGameVersions(
		c.Request().Context(),
		values.NewGameIDFromUUID(gameID),
		&service.GetGameVersionsParams{
			Limit:  limit,
			Offset: offset,
		},
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
			Id:          openapi.GameVersionID(version.GameVersion.GetID()),
			Name:        string(version.GameVersion.GetName()),
			Description: string(version.GameVersion.GetDescription()),
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
