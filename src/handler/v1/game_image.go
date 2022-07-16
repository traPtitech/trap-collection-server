package v1

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"

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

func (gi *GameImage) PostImage(c echo.Context, strGameID string, image multipart.File) error {
	ctx := c.Request().Context()

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

func (gi *GameImage) GetImage(c echo.Context, strGameID string) error {
	ctx := c.Request().Context()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	tmpURL, err := gi.gameImageService.GetGameImage(ctx, gameID)
	if errors.Is(err, service.ErrNoGameImage) {
		return echo.NewHTTPError(http.StatusNotFound, "no image")
	}
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if err != nil {
		log.Printf("error: failed to get image: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get image")
	}

	c.Response().Header().Set(echo.HeaderLocation, (*url.URL)(tmpURL).String())

	return echo.NewHTTPError(http.StatusSeeOther, fmt.Sprintf("redirect to %s", (*url.URL)(tmpURL).String()))
}
