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

type GameImage struct {
	gameImageUnimplemented
	gameImageService service.GameImageV2
}

func NewGameImage(gameImageService service.GameImageV2) *GameImage {
	return &GameImage{
		gameImageService: gameImageService,
	}
}

// gameImageUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameImageUnimplemented interface {
	// ゲーム画像一覧の取得
	// (POST /games/{gameID}/images)
	PostGameImage(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム画像のバイナリの取得
	// (GET /games/{gameID}/images/{gameImageID})
	GetGameImage(ctx echo.Context, gameID openapi.GameIDInPath, gameImageID openapi.GameImageIDInPath) error
	// ゲーム画像のメタ情報の取得
	// (GET /games/{gameID}/images/{gameImageID}/meta)
	GetGameImageMeta(ctx echo.Context, gameID openapi.GameIDInPath, gameImageID openapi.GameImageIDInPath) error
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
