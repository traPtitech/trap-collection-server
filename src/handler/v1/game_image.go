package v1

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameImage struct {
	gameImageService service.GameImage
}

func NewGameImage(gameImageService service.GameImage) *GameImage {
	return &GameImage{
		gameImageService: gameImageService,
	}
}

func (gi *GameImage) PostImage(strGameID string, image multipart.File) error {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	err = gi.gameImageService.SaveGameImage(ctx, image, gameID)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if errors.Is(err, service.ErrInvalidFormat) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image format")
	}
	if err != nil {
		log.Printf("error: failed to save image: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save image")
	}

	return nil
}
