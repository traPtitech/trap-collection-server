package v1

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

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
	session, err := getSession(c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	authSession, err := g.session.getAuthSession(session)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth session")
	}

	name := values.NewGameName(newGame.Name)
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

	description := values.NewGameDescription(newGame.Description)

	game, err := g.gameService.CreateGame(
		c.Request().Context(),
		authSession,
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

	var gameVersion *openapi.GameVersion
	if game.LatestVersion != nil {
		gameVersion = &openapi.GameVersion{
			Id:          uuid.UUID(game.LatestVersion.GetID()).String(),
			Name:        string(game.LatestVersion.GetName()),
			Description: string(game.LatestVersion.GetDescription()),
			CreatedAt:   game.LatestVersion.GetCreatedAt(),
		}
	}

	return &openapi.Game{
		Id:          uuid.UUID(game.Game.GetID()).String(),
		Name:        string(game.Game.GetName()),
		Description: string(game.Game.GetDescription()),
		CreatedAt:   game.Game.GetCreatedAt(),
		Version:     gameVersion,
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

func (g *Game) GetGames(strAll string, c echo.Context) ([]*openapi.Game, error) {
	var isAll bool
	if len(strAll) != 0 {
		isAll = false
	} else {
		var err error
		isAll, err = strconv.ParseBool(strAll)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "all is invalid")
		}
	}

	var games []*service.GameInfo
	var err error
	if isAll {
		games, err = g.gameService.GetGames(c.Request().Context())
		if err != nil {
			log.Printf("error: failed to get games: %v\n", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get games")
		}
	} else {
		session, err := getSession(c)
		if err != nil {
			log.Printf("error: failed to get session: %v\n", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
		}

		authSession, err := g.session.getAuthSession(session)
		if err != nil {
			// middlewareでログイン済みなことは確認しているので、ここではエラーになりえないはず
			log.Printf("error: failed to get auth session: %v\n", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError)
		}

		games, err = g.gameService.GetMyGames(c.Request().Context(), authSession)
		if err != nil {
			log.Printf("error: failed to get latest games: %v\n", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get latest games")
		}
	}

	gameInfos := make([]*openapi.Game, 0, len(games))
	for _, game := range games {
		var gameVersion *openapi.GameVersion
		if game.LatestVersion != nil {
			gameVersion = &openapi.GameVersion{
				Id:          uuid.UUID(game.LatestVersion.GetID()).String(),
				Name:        string(game.LatestVersion.GetName()),
				Description: string(game.LatestVersion.GetDescription()),
				CreatedAt:   game.LatestVersion.GetCreatedAt(),
			}
		}

		gameInfos = append(gameInfos, &openapi.Game{
			Id:          uuid.UUID(game.Game.GetID()).String(),
			Name:        string(game.Game.GetName()),
			Description: string(game.Game.GetDescription()),
			CreatedAt:   game.Game.GetCreatedAt(),
			Version:     gameVersion,
		})
	}

	return gameInfos, nil
}

func (g *Game) DeleteGames(strGameID string) error {
	ctx := context.Background()

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	err = g.gameService.DeleteGame(ctx, gameID)
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusBadRequest, "no game")
	}
	if err != nil {
		log.Printf("error: failed to delete game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete game")
	}

	return nil
}
