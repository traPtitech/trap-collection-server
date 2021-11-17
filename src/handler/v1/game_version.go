package v1

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVersion struct {
	gameVersionService service.GameVersion
}

func NewGameVersion(gameVersionService service.GameVersion) *GameVersion {
	return &GameVersion{
		gameVersionService: gameVersionService,
	}
}

func (gv *GameVersion) PostGameVersion(strGameID string, newGameVersion *openapi.NewGameVersion) (*openapi.GameVersion, error) {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	name := values.NewGameVersionName(newGameVersion.Name)
	err = name.Validate()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid game version name: %s", err.Error()))
	}

	description := values.NewGameVersionDescription(newGameVersion.Description)

	gameVersion, err := gv.gameVersionService.CreateGameVersion(ctx, gameID, name, description)
	if errors.Is(err, service.ErrInvalidGameID) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if err != nil {
		log.Printf("error: failed to create game version: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to create game version")
	}

	return &openapi.GameVersion{
		Id:          uuid.UUID(gameVersion.GetID()).String(),
		Name:        string(gameVersion.GetName()),
		Description: string(gameVersion.GetDescription()),
		CreatedAt:   gameVersion.GetCreatedAt(),
	}, nil
}
