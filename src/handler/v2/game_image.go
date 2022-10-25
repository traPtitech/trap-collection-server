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

type GameImage struct {
	gameImageService service.GameImageV2
}

func NewGameImage(gameImageService service.GameImageV2) *GameImage {
	return &GameImage{
		gameImageService: gameImageService,
	}
}

// ゲーム画像一覧の取得
// (GET /games/{gameID}/images)
func (gameImage *GameImage) GetGameImages(c echo.Context, gameID openapi.GameIDInPath) error {
	images, err := gameImage.gameImageService.GetGameImages(c.Request().Context(), values.NewGameIDFromUUID(gameID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if err != nil {
		log.Printf("error: failed to get game images: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game images")
	}

	resImages := make([]openapi.GameImage, 0, len(images))
	for _, image := range images {
		var mime openapi.GameImageMime
		switch image.GetType() {
		case values.GameImageTypeJpeg:
			mime = openapi.Imagejpeg
		case values.GameImageTypePng:
			mime = openapi.Imagepng
		case values.GameImageTypeGif:
			mime = openapi.Imagegif
		default:
			log.Printf("error: unknown game image type: %v\n", image.GetType())
			return echo.NewHTTPError(http.StatusInternalServerError, "unknown game image type")
		}

		resImages = append(resImages, openapi.GameImage{
			Id:        openapi.GameImageID(image.GetID()),
			Mime:      mime,
			CreatedAt: image.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, resImages)
}

// ゲームファイルの作成
// (POST /games/{gameID}/images)
func (gameImage *GameImage) PostGameImage(c echo.Context, gameID openapi.GameIDInPath) error {
	header, err := c.FormFile("content")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid file")
	}
	file, err := header.Open()
	if err != nil {
		log.Printf("error: failed to open file: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open file")
	}
	defer file.Close()

	image, err := gameImage.gameImageService.SaveGameImage(c.Request().Context(), file, values.NewGameIDFromUUID(gameID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidFormat) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image type")
	}
	if err != nil {
		log.Printf("error: failed to save game image: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save game image")
	}

	var mime openapi.GameImageMime
	switch image.GetType() {
	case values.GameImageTypeJpeg:
		mime = openapi.Imagejpeg
	case values.GameImageTypePng:
		mime = openapi.Imagepng
	case values.GameImageTypeGif:
		mime = openapi.Imagegif
	default:
		log.Printf("error: unknown game image type: %v\n", image.GetType())
		return echo.NewHTTPError(http.StatusInternalServerError, "unknown game image type")
	}

	return c.JSON(http.StatusCreated, openapi.GameImage{
		Id:        openapi.GameImageID(image.GetID()),
		Mime:      mime,
		CreatedAt: image.GetCreatedAt(),
	})
}

// ゲーム画像のバイナリの取得
// (GET /games/{gameID}/images/{gameImageID})
func (gameImage *GameImage) GetGameImage(c echo.Context, gameID openapi.GameIDInPath, gameImageID openapi.GameImageIDInPath) error {
	tmpURL, err := gameImage.gameImageService.GetGameImage(c.Request().Context(), values.NewGameIDFromUUID(gameID), values.GameImageIDFromUUID(gameImageID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidGameImageID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameImageID")
	}
	if err != nil {
		log.Printf("error: failed to get game image: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game image")
	}

	return c.Redirect(http.StatusSeeOther, (*url.URL)(tmpURL).String())
}

// ゲーム画像のメタ情報の取得
// (GET /games/{gameID}/images/{gameImageID}/meta)
func (gameImage *GameImage) GetGameImageMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameImageID openapi.GameImageIDInPath) error {
	image, err := gameImage.gameImageService.GetGameImageMeta(ctx.Request().Context(), values.NewGameIDFromUUID(gameID), values.GameImageIDFromUUID(gameImageID))
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameID")
	}
	if errors.Is(err, service.ErrInvalidGameImageID) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid gameImageID")
	}
	if err != nil {
		log.Printf("error: failed to get game image meta: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game image meta")
	}

	var mime openapi.GameImageMime
	switch image.GetType() {
	case values.GameImageTypeJpeg:
		mime = openapi.Imagejpeg
	case values.GameImageTypePng:
		mime = openapi.Imagepng
	case values.GameImageTypeGif:
		mime = openapi.Imagegif
	default:
		log.Printf("error: unknown game image type: %v\n", image.GetType())
		return echo.NewHTTPError(http.StatusInternalServerError, "unknown game image type")
	}

	return ctx.JSON(http.StatusOK, openapi.GameImage{
		Id:        openapi.GameImageID(image.GetID()),
		Mime:      mime,
		CreatedAt: image.GetCreatedAt(),
	})
}
