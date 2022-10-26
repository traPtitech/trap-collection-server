package v1

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameVersion struct {
	featureWrite       bool
	gameVersionService service.GameVersion
}

func NewGameVersion(appConf config.App, gameVersionService service.GameVersion) *GameVersion {
	return &GameVersion{
		featureWrite:       appConf.FeatureV1Write(),
		gameVersionService: gameVersionService,
	}
}

func (gv *GameVersion) PostGameVersion(c echo.Context, strGameID string, newGameVersion *openapi.NewGameVersion) (*openapi.GameVersion, error) {
	if !gv.featureWrite {
		return nil, echo.NewHTTPError(http.StatusForbidden, "v1 write is disabled")
	}

	ctx := c.Request().Context()

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

func (gv *GameVersion) GetGameVersion(c echo.Context, strGameID string) ([]*openapi.GameVersion, error) {
	ctx := c.Request().Context()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	gameVersions, err := gv.gameVersionService.GetGameVersions(ctx, gameID)
	if errors.Is(err, service.ErrInvalidGameID) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if err != nil {
		log.Printf("error: failed to get game versions: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get game versions")
	}

	apiGameVersions := make([]*openapi.GameVersion, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		apiGameVersions = append(apiGameVersions, &openapi.GameVersion{
			Id:          uuid.UUID(gameVersion.GetID()).String(),
			Name:        string(gameVersion.GetName()),
			Description: string(gameVersion.GetDescription()),
			CreatedAt:   gameVersion.GetCreatedAt(),
		})
	}

	return apiGameVersions, nil
}
