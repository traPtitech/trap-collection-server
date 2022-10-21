package v2

import (
	"errors"
	"log"
	"net/http"

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

// gameFileUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameFileUnimplemented interface {
	// ゲームファイルの作成
	// (GET /games/{gameID}/files)
	GetGameFiles(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームファイル一覧の取得
	// (POST /games/{gameID}/files)
	PostGameFile(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームファイルのバイナリの取得
	// (GET /games/{gameID}/files/{gameFileID})
	GetGameFile(ctx echo.Context, gameID openapi.GameIDInPath, gameFileID openapi.GameFileIDInPath) error
	// ゲームファイルのメタ情報の取得
	// (GET /games/{gameID}/files/{gameFileID}/meta)
	GetGameFileMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameFileID openapi.GameFileIDInPath) error
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

		resFiles = append(resFiles, openapi.GameFile{
			Id:         openapi.GameFileID(file.GetID()),
			Type:       fileType,
			EntryPoint: string(file.GetEntryPoint()),
			Md5:        string(file.GetHash()),
			CreatedAt:  file.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, files)
}

// ゲームファイルの作成
// (POST /games/{gameID}/files)
func (gameFile GameFile) PostGameFile(c echo.Context, gameID openapi.GameIDInPath) error {
	headerFile, err := c.FormFile("content")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid file")
	}
	headerEntryPoint := c.FormValue("entryPoint")
	headerFileType := c.FormValue("fileType")

	file, err := headerFile.Open()
	if err != nil {
		log.Printf("error: failed to open file: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open file")
	}
	defer file.Close()

	entryPoint := values.NewGameFileEntryPoint(headerEntryPoint)
	var fileType values.GameFileType
	switch headerFileType {
	case "jar":
		fileType = values.GameFileTypeJar
	case "windows":
		fileType = values.GameFileTypeWindows
	case "darwin":
		fileType = values.GameFileTypeMac
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "file type is unknown")
	}

	savedFile, err := gameFile.gameFileService.SaveGameFile(c.Request().Context(), file, values.NewGameIDFromUUID(gameID), fileType, entryPoint)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if err != nil {
		log.Printf("error: failed to save game image: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save game image")
	}

	return c.JSON(http.StatusCreated, openapi.GameFile{
		Id:         openapi.GameFileID(savedFile.GetID()),
		Type:       openapi.GameFileType(headerFileType),
		EntryPoint: openapi.GameFileEntryPoint(savedFile.GetEntryPoint()),
		Md5:        openapi.GameFileMd5(savedFile.GetHash()),
		CreatedAt:  savedFile.GetCreatedAt(),
	})
}
