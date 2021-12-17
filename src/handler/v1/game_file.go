package v1

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameFile struct {
	gameFileService service.GameFile
}

func NewGameFile(gameFileService service.GameFile) *GameFile {
	return &GameFile{
		gameFileService: gameFileService,
	}
}

func (gf *GameFile) PostFile(strGameID string, entryPoint string, file multipart.File, strFileType string) (*openapi.GameFile, error) {
	ctx := context.Background()

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

	gameFile, err := gf.gameFileService.SaveGameFile(
		ctx,
		file,
		gameID,
		fileType,
		values.NewGameFileEntryPoint(entryPoint),
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
		EntryPoint: entryPoint,
	}, nil
}
