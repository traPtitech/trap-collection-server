package v1

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameFile struct {
	featureWrite    bool
	gameFileService service.GameFile
}

func NewGameFile(appConf config.App, gameFileService service.GameFile) *GameFile {
	return &GameFile{
		featureWrite:    appConf.FeatureV1Write(),
		gameFileService: gameFileService,
	}
}

func (gf *GameFile) PostFile(c echo.Context, strGameID string, strEntryPoint string, strFileType string, file multipart.File) (*openapi.GameFile, error) {
	if !gf.featureWrite {
		return nil, echo.NewHTTPError(http.StatusForbidden, "v1 write is disabled")
	}

	ctx := c.Request().Context()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	var fileType values.GameFileType
	switch strFileType {
	case "jar":
		fileType = values.GameFileTypeJar
	case "windows":
		fileType = values.GameFileTypeWindows
	case "mac":
		fileType = values.GameFileTypeMac
	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid file type")
	}

	entryPoint := values.NewGameFileEntryPoint(strEntryPoint)
	err = entryPoint.Validate()
	if errors.Is(err, values.ErrGameFileEntryPointEmpty) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "entry point must not be empty")
	}
	if err != nil {
		log.Printf("error: failed to validate entry point: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to validate entry point")
	}

	gameFile, err := gf.gameFileService.SaveGameFile(
		ctx,
		file,
		gameID,
		fileType,
		entryPoint,
	)
	if errors.Is(err, service.ErrInvalidGameID) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if errors.Is(err, service.ErrNoGameVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no game version")
	}
	if errors.Is(err, service.ErrGameFileAlreadyExists) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "game file already exists")
	}
	if err != nil {
		log.Printf("error: failed to save file: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to save file")
	}

	return &openapi.GameFile{
		Id:         uuid.UUID(gameFile.GetID()).String(),
		Type:       strFileType,
		EntryPoint: string(gameFile.GetEntryPoint()),
	}, nil
}

func (gf *GameFile) GetGameFile(c echo.Context, strGameID string, strOperatingSystem string) (io.ReadCloser, error) {
	ctx := c.Request().Context()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	var envOS values.LauncherEnvironmentOS
	switch strOperatingSystem {
	case "win32":
		envOS = values.LauncherEnvironmentOSWindows
	case "darwin":
		envOS = values.LauncherEnvironmentOSMac
	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid operating system")
	}

	tmpURL, _, err := gf.gameFileService.GetGameFile(
		ctx,
		gameID,
		values.NewLauncherEnvironment(envOS),
	)
	if errors.Is(err, service.ErrInvalidGameID) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if errors.Is(err, service.ErrNoGameVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no game version")
	}
	if errors.Is(err, service.ErrNoGameFile) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no game file")
	}
	if err != nil {
		log.Printf("error: failed to get game file: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get game file")
	}

	c.Response().Header().Set(echo.HeaderLocation, (*url.URL)(tmpURL).String())

	return nil, echo.NewHTTPError(http.StatusSeeOther, fmt.Sprintf("redirect to %s", (*url.URL)(tmpURL).String()))
}
