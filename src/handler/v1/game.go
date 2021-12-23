package v1

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Game struct {
	session     *Session
	gameService service.Game
}

func NewGame(
	session *Session,
	gameService service.Game,
) *Game {
	return &Game{
		session:     session,
		gameService: gameService,
	}
}

func (g *Game) PostGame(newGame *openapi.NewGame, c echo.Context) (*openapi.GameInfo, error) {
	name := values.NewGameName(newGame.Name)
	err := name.Validate()
	if errors.Is(err, values.ErrGameNameEmpty) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "game name is empty")
	}
	if errors.Is(err, values.ErrGameNameTooLong) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "game name is too long")
	}
	if err != nil {
		log.Printf("error: failed to validate game name: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to validate game name")
	}

	description := values.NewGameDescription(newGame.Description)

	game, err := g.gameService.CreateGame(
		c.Request().Context(),
		name,
		description,
	)
	if err != nil {
		log.Printf("error: failed to create game: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to create game")
	}

	return &openapi.GameInfo{
		Id:          uuid.UUID(game.GetID()).String(),
		Name:        string(game.GetName()),
		Description: string(game.GetDescription()),
		CreatedAt:   game.GetCreatedAt(),
	}, nil
}

func (g *Game) GetGame(strGameID string) (*openapi.Game, error) {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	game, err := g.gameService.GetGame(ctx, gameID)
	if errors.Is(err, service.ErrNoGame) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no game")
	}
	if err != nil {
		log.Printf("error: failed to get game: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get game")
	}

	return &openapi.Game{
		Id:          uuid.UUID(game.Game.GetID()).String(),
		Name:        string(game.Game.GetName()),
		Description: string(game.Game.GetDescription()),
		CreatedAt:   game.Game.GetCreatedAt(),
		Version: &openapi.GameVersion{
			Id:          uuid.UUID(game.LatestVersion.GetID()).String(),
			Name:        string(game.LatestVersion.GetName()),
			Description: string(game.LatestVersion.GetDescription()),
			CreatedAt:   game.LatestVersion.GetCreatedAt(),
		},
	}, nil
}

func (g *Game) PutGame(strGameID string, gameMeta *openapi.NewGame) (*openapi.GameInfo, error) {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	name := values.NewGameName(gameMeta.Name)
	err = name.Validate()
	if errors.Is(err, values.ErrGameNameEmpty) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "game name is empty")
	}
	if errors.Is(err, values.ErrGameNameTooLong) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "game name is too long")
	}
	if err != nil {
		log.Printf("error: failed to validate game name: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to validate game name")
	}

	description := values.NewGameDescription(gameMeta.Description)

	game, err := g.gameService.UpdateGame(ctx, gameID, name, description)
	if errors.Is(err, service.ErrNoGame) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no game")
	}
	if err != nil {
		log.Printf("error: failed to update game: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to update game")
	}

	return &openapi.GameInfo{
		Id:          uuid.UUID(game.GetID()).String(),
		Name:        string(game.GetName()),
		Description: string(game.GetDescription()),
		CreatedAt:   game.GetCreatedAt(),
	}, nil
}
