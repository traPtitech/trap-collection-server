package v2

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/mazrean/formstream"
	echoform "github.com/mazrean/formstream/echo"
	"github.com/traPtitech/trap-collection-server/src/domain"
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

		resFiles = append(resFiles, openapi.GameFile{
			Id:         openapi.GameFileID(file.GetID()),
			Type:       fileType,
			EntryPoint: string(file.GetEntryPoint()),
			Md5:        hex.EncodeToString(file.GetHash()),
			CreatedAt:  file.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, resFiles)
}

// ゲームファイルの作成
// (POST /games/{gameID}/files)
func (gameFile GameFile) PostGameFile(c echo.Context, gameID openapi.GameIDInPath) error {
	parser, err := echoform.NewParser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	var (
		noContent      = true
		headerFileType string
		savedFile      *domain.GameFile
	)
	err = parser.Register("content", func(r io.Reader, _ formstream.Header) error {
		noContent = false

		headerEntryPoint, _, _ := parser.Value("entryPoint")
		headerFileType, _, _ = parser.Value("type")

		entryPoint := values.NewGameFileEntryPoint(headerEntryPoint)
		if err := entryPoint.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid entry point")
		}

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

		var err error
		savedFile, err = gameFile.gameFileService.SaveGameFile(c.Request().Context(), r, values.NewGameIDFromUUID(gameID), fileType, entryPoint)
		if errors.Is(err, service.ErrInvalidGameID) {
			return echo.NewHTTPError(http.StatusNotFound, "gameID not found")
		}
		if errors.Is(err, service.ErrNotZipFile) {
			return echo.NewHTTPError(http.StatusBadRequest, "only zip file is allowed")
		}
		if errors.Is(err, service.ErrInvalidEntryPoint) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid entry point")
		}
		if err != nil {
			log.Printf("error: failed to save game file: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to save game file")
		}

		return nil
	}, formstream.WithRequiredPart("entryPoint"), formstream.WithRequiredPart("type"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register parser")
	}

	if err := parser.Parse(); err != nil {
		return err
	}

	if _, _, ok := parser.Value("entryPoint"); !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "no entry point")
	}

	if _, _, ok := parser.Value("type"); !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "no file type")
	}

	if noContent {
		return echo.NewHTTPError(http.StatusBadRequest, "no content")
	}

	return c.JSON(http.StatusCreated, openapi.GameFile{
		Id:         openapi.GameFileID(savedFile.GetID()),
		Type:       openapi.GameFileType(headerFileType),
		EntryPoint: openapi.GameFileEntryPoint(savedFile.GetEntryPoint()),
		Md5:        openapi.GameFileMd5(hex.EncodeToString(savedFile.GetHash())),
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
		fileType = openapi.Jar
	case values.GameFileTypeWindows:
		fileType = openapi.Win32
	case values.GameFileTypeMac:
		fileType = openapi.Darwin
	default:
		log.Printf("error: unknown game file type: %v\n", file.GetFileType())
		return echo.NewHTTPError(http.StatusInternalServerError, "unknown game file type")
	}

	return ctx.JSON(http.StatusOK, openapi.GameFile{
		Id:         openapi.GameFileID(file.GetID()),
		Type:       fileType,
		EntryPoint: openapi.GameFileEntryPoint(file.GetEntryPoint()),
		Md5:        openapi.GameFileMd5(hex.EncodeToString(file.GetHash())),
		CreatedAt:  file.GetCreatedAt(),
	})
}
