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
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVideo struct {
	featureWrite     bool
	gameVideoService service.GameVideo
}

func NewGameVideo(appConf config.App, gameVideoService service.GameVideo) *GameVideo {
	return &GameVideo{
		featureWrite:     appConf.FeatureV1Write(),
		gameVideoService: gameVideoService,
	}
}

func (gv *GameVideo) PostVideo(c echo.Context, strGameID string, video multipart.File) error {
	if !gv.featureWrite {
		return echo.NewHTTPError(http.StatusForbidden, "v1 write is disabled")
	}

	ctx := c.Request().Context()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	err = gv.gameVideoService.SaveGameVideo(ctx, video, gameID)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if errors.Is(err, service.ErrInvalidFormat) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid video format")
	}
	if err != nil {
		log.Printf("error: failed to save video: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save video")
	}

	return nil
}

func (gv *GameVideo) GetVideo(c echo.Context, strGameID string) error {
	ctx := c.Request().Context()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	tmpURL, err := gv.gameVideoService.GetGameVideo(ctx, gameID)
	if errors.Is(err, service.ErrNoGameVideo) {
		return echo.NewHTTPError(http.StatusNotFound, "no video")
	}
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if err != nil {
		log.Printf("error: failed to get video: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get video")
	}

	c.Response().Header().Set(echo.HeaderLocation, (*url.URL)(tmpURL).String())

	return echo.NewHTTPError(http.StatusSeeOther, fmt.Sprintf("redirect to %s", (*url.URL)(tmpURL).String()))
}
