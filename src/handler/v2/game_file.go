package v2

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameFile struct {
	gameFileService service.GameFileV2
}

func NewGameFile(gameFileService service.GameFileV2) *GameFile {
	return &GameFile{
		gameFileService: gameFileService,
	}
}

// ゲームファイル一覧の取得
// (GET /games/{gameID}/files)
func (gameFile GameFile) GetGameFiles(c echo.Context, gameID openapi.GameIDInPath) error {
	files, err := gameFile.gameFileService.GetGameFiles(c.Request().Context(), values.NewGameIDFromUUID(gameID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if err != nil {
		log.Printf("error: failed to get game files: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game files")
	}

	resFiles := make([]openapi.GameFile, 0, len(files))
	for _, file := range files {
		var fileType openapi.GameFileType
		switch file.GetFileType() {
		case values.GameFileTypeJar:
			fileType = openapi.Jar
		case values.GameFileTypeWindows:
			fileType = openapi.Win32
		case values.GameFileTypeMac:
			fileType = openapi.Darwin
		default:
			log.Printf("error: unknown game file type: %v\n", file.GetFileType())
			return echo.NewHTTPError(http.StatusInternalServerError, "unknown game file type")
		}

		log.Printf("res: %v ->%v", string(file.GetHash()), []byte(file.GetHash())) //TODO:後で消す
		resFiles = append(resFiles, openapi.GameFile{
			Id:         openapi.GameFileID(file.GetID()),
			Type:       fileType,
			EntryPoint: string(file.GetEntryPoint()),
			Md5:        string(file.GetHash()),
			CreatedAt:  file.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, resFiles)
}

// ゲームファイルの作成
// (POST /games/{gameID}/files)
func (gameFile GameFile) PostGameFile(c echo.Context, gameID openapi.GameIDInPath) error {
	headerFile, err := c.FormFile("content")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid file")
	}
	headerEntryPoint := c.FormValue("entryPoint")
	headerFileType := c.FormValue("type")

	if headerEntryPoint == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "entry point is empty")
	}
	if headerFileType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "file type is empty")
	}

	file, err := headerFile.Open()
	if err != nil {
		log.Printf("error: failed to open file: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open file")
	}
	defer file.Close()

	entryPoint := values.NewGameFileEntryPoint(headerEntryPoint)
	var fileType values.GameFileType
	switch openapi.GameFileType(headerFileType) {
	case openapi.Jar:
		fileType = values.GameFileTypeJar
	case openapi.Win32:
		fileType = values.GameFileTypeWindows
	case openapi.Darwin:
		fileType = values.GameFileTypeMac
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "file type is unknown")
	}

	savedFile, err := gameFile.gameFileService.SaveGameFile(c.Request().Context(), file, values.NewGameIDFromUUID(gameID), fileType, entryPoint)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if err != nil {
		log.Printf("error: failed to save game file: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save game file")
	}

	return c.JSON(http.StatusCreated, openapi.GameFile{
		Id:         openapi.GameFileID(savedFile.GetID()),
		Type:       openapi.GameFileType(headerFileType),
		EntryPoint: openapi.GameFileEntryPoint(savedFile.GetEntryPoint()),
		Md5:        openapi.GameFileMd5(savedFile.GetHash()),
		CreatedAt:  savedFile.GetCreatedAt(),
	})
}

// ゲームファイルのバイナリの取得
// (GET /games/{gameID}/files/{gameFileID})
func (gameFile GameFile) GetGameFile(c echo.Context, gameID openapi.GameIDInPath, gameFileID openapi.GameFileIDInPath) error {
	tmpURL, err := gameFile.gameFileService.GetGameFile(c.Request().Context(), values.NewGameIDFromUUID(gameID), values.NewGameFileIDFromUUID(gameFileID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidGameFileID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameFileID")
	}
	if err != nil {
		log.Printf("error: failed to get game file: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game file")
	}

	return c.Redirect(http.StatusSeeOther, (*url.URL)(tmpURL).String())
}

// ゲームファイルのメタ情報の取得
// (GET /games/{gameID}/files/{gameFileID}/meta)
func (gameFile GameFile) GetGameFileMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameFileID openapi.GameFileIDInPath) error {
	file, err := gameFile.gameFileService.GetGameFileMeta(ctx.Request().Context(), values.NewGameIDFromUUID(gameID), values.NewGameFileIDFromUUID(gameFileID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidGameFileID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameFileID")
	}
	if err != nil {
		log.Printf("error: failed to get game file meta: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game file meta")
	}

	var fileType openapi.GameFileType
	switch file.GetFileType() {
	case values.GameFileTypeJar:
		fileType = openapi.GameFileType("jar")
	case values.GameFileTypeWindows:
		fileType = openapi.GameFileType("windows")
	case values.GameFileTypeMac:
		fileType = openapi.GameFileType("darwin")
	default:
		log.Printf("error: unknown game file type: %v\n", file.GetFileType())
		return echo.NewHTTPError(http.StatusInternalServerError, "unknown game file type")
	}

	return ctx.JSON(http.StatusOK, openapi.GameFile{
		Id:         openapi.GameFileID(file.GetID()),
		Type:       fileType,
		EntryPoint: openapi.GameFileEntryPoint(file.GetEntryPoint()),
		Md5:        openapi.GameFileMd5(file.GetHash()),
		CreatedAt:  file.GetCreatedAt(),
	})
}
