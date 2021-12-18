package v1

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameURL struct {
	gameURLService service.GameURL
}

func NewGameURL(gameURLService service.GameURL) *GameURL {
	return &GameURL{
		gameURLService: gameURLService,
	}
}

func (gu *GameURL) PostURL(strGameID string, newGameURL *openapi.NewGameUrl) (*openapi.GameUrl, error) {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	urlLink, err := url.Parse(newGameURL.Url)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid url")
	}

	gameURL, err := gu.gameURLService.SaveGameURL(
		ctx,
		gameID,
		values.NewGameURLLink(urlLink),
	)
	if errors.Is(err, service.ErrInvalidGameID) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if errors.Is(err, service.ErrNoGameVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no game version")
	}
	if errors.Is(err, service.ErrGameURLAlreadyExists) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "game url already exists")
	}
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to save game url")
	}

	return &openapi.GameUrl{
		Id:  uuid.UUID(gameURL.GetID()).String(),
		Url: (*url.URL)(gameURL.GetLink()).String(),
	}, nil
}
