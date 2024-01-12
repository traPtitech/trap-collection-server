package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Game struct {
	session     *Session
	gameService service.GameV2
}

func NewGame(session *Session, gameService service.GameV2) *Game {
	return &Game{
		session:     session,
		gameService: gameService,
	}
}

// ゲーム一覧の取得
// (GET /games)
func (g *Game) GetGames(ctx echo.Context, params openapi.GetGamesParams) error {
	var isAll bool
	if params.All != nil {
		isAll = *params.All
	} else {
		isAll = true
	}

	var limit int
	var offset int
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	var games []*domain.Game
	var gameNumber int
	var err error
	if isAll {
		gameNumber, games, err = g.gameService.GetGames(ctx.Request().Context(), limit, offset)
		if err != nil {
			log.Printf("error: failed to get games: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get games")
		}
	} else {
		session, err := g.session.get(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "no session")
		}
		authSession, err := g.session.getAuthSession(session)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "no auth session")
		}

		gameNumber, games, err = g.gameService.GetMyGames(ctx.Request().Context(), authSession, limit, offset)
		if err != nil {
			log.Printf("error: failed to get games: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get my games")
		}
	}

	responseGames := make([]openapi.GameInfo, 0, len(games))
	for _, game := range games {
		responseGame := openapi.GameInfo{
			Name:        string(game.GetName()),
			Id:          uuid.UUID(game.GetID()),
			Description: string(game.GetDescription()),
			CreatedAt:   game.GetCreatedAt(),
		}
		responseGames = append(responseGames, responseGame)
	}

	//TODO: ビルドを通すためにいったん仮の配列を返している。全体が正しく動くよう修正する必要がある。
	gamesWithGenres := make([]openapi.GameInfoWithGenres, 0, len(responseGames))
	for _, game := range responseGames {
		gameWithGenre := openapi.GameInfoWithGenres{
			Name:        game.Name,
			Id:          game.Id,
			Description: game.Description,
			CreatedAt:   game.CreatedAt,
			Genres:      &[]openapi.GameGenreName{},
		}
		gamesWithGenres = append(gamesWithGenres, gameWithGenre)
	}

	res := openapi.GetGamesResponse{
		Games: gamesWithGenres,
		Num:   gameNumber,
	}

	return ctx.JSON(http.StatusOK, res)
}

// ゲームの追加
// (POST /games)
func (g *Game) PostGame(ctx echo.Context) error {
	session, err := g.session.get(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no session")
	}
	authSession, err := g.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no auth session")
	}

	req := &openapi.PostGameJSONRequestBody{}
	err = ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	gameName := values.NewGameName(req.Name)
	err = gameName.Validate()
	if errors.Is(err, values.ErrGameNameEmpty) {
		return echo.NewHTTPError(http.StatusBadRequest, "game name is empty")
	}
	if errors.Is(err, values.ErrGameNameTooLong) {
		return echo.NewHTTPError(http.StatusBadRequest, "game name is too long")
	}
	if err != nil {
		log.Printf("error: failed to validate game name: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to validate game name")
	}

	var owners []values.TraPMemberName
	if req.Owners != nil {
		owners = make([]values.TraPMemberName, 0, len(*req.Owners))
		for _, reqOwner := range *req.Owners {
			owners = append(owners, values.NewTrapMemberName(reqOwner))
		}
	}

	var maintainers []values.TraPMemberName
	if req.Maintainers != nil {
		maintainers = make([]values.TraPMemberName, 0, len(*req.Maintainers))
		for _, reqMaintainer := range *req.Maintainers {
			maintainers = append(maintainers, values.NewTrapMemberName(reqMaintainer))
		}
	}

	var visibility values.GameVisibility
	switch req.Visibility {
	case openapi.Public:
		visibility = values.GameVisibilityTypePublic
	case openapi.Limited:
		visibility = values.GameVisibilityTypeLimited
	case openapi.Private:
		visibility = values.GameVisibilityTypePrivate
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid visibility")
	}

	var genreNames []values.GameGenreName
	if req.Genres != nil {
		genreNames = make([]values.GameGenreName, 0, len(*req.Genres))
		for i := range *req.Genres {
			genreName := values.NewGameGenreName((*req.Genres)[i])
			err := genreName.Validate()
			if errors.Is(err, values.ErrGameGenreNameEmpty) {
				return echo.NewHTTPError(http.StatusBadRequest, "game genre name is empty")
			}
			if errors.Is(err, values.ErrGameGenreNameTooLong) {
				return echo.NewHTTPError(http.StatusBadRequest, "game genre name is too long")
			}
			if err != nil {
				log.Printf("failed to validate game genre name: %v\n", genreName)
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to validate game genre name")
			}

			genreNames = append(genreNames, genreName)
		}
	}

	gameInfo, err := g.gameService.CreateGame(
		ctx.Request().Context(),
		authSession,
		gameName,
		values.GameDescription(req.Description),
		visibility,
		owners,
		maintainers,
		genreNames,
	)

	if errors.Is(err, service.ErrOverlapInOwners) {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to add owners")
	}
	if errors.Is(err, service.ErrOverlapInMaintainers) {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to add maintainers")
	}
	if errors.Is(err, service.ErrOverlapBetweenOwnersAndMaintainers) {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to add owners and maintainers")
	}
	if errors.Is(err, service.ErrDuplicateGameGenre) {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to add game genre")
	}
	if err != nil {
		log.Printf("error: failed to create game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create game")
	}

	resOwners := make([]string, 0, len(gameInfo.Owners))
	for _, owner := range gameInfo.Owners {
		resOwners = append(resOwners, string(owner.GetName()))
	}

	resMaintainers := make([]string, 0, len(gameInfo.Maintainers))
	for _, maintainer := range gameInfo.Maintainers {
		resMaintainers = append(resMaintainers, string(maintainer.GetName()))
	}

	var resGameGenreNames []openapi.GameGenreName
	if gameInfo.Genres != nil && len(gameInfo.Genres) != 0 { // ジャンルが無い場合はnilにする
		resGameGenreNames = make([]openapi.GameGenreName, 0, len(gameInfo.Genres))
		for _, genre := range gameInfo.Genres {
			resGameGenreNames = append(resGameGenreNames, openapi.GameGenreName(genre.GetName()))
		}
	}

	res := openapi.Game{
		Name:        string(gameInfo.Game.GetName()),
		Id:          uuid.UUID(gameInfo.Game.GetID()),
		Description: string(gameInfo.Game.GetDescription()),
		CreatedAt:   gameInfo.Game.GetCreatedAt(),
		Owners:      resOwners,
		Maintainers: &resMaintainers,
		Genres:      &resGameGenreNames,
	}

	return ctx.JSON(http.StatusCreated, res)
}

// ゲームの削除
// (DELETE /games/{gameID})
func (g *Game) DeleteGame(ctx echo.Context, gameID openapi.GameIDInPath) error {
	err := g.gameService.DeleteGame(ctx.Request().Context(), values.GameID(gameID))
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	} else if err != nil {
		log.Printf("error: failed to delete game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.NoContent(http.StatusOK)
}

// ゲーム情報の取得
// (GET /games/{gameID})
func (g *Game) GetGame(ctx echo.Context, gameID openapi.GameIDInPath) error {
	session, err := g.session.get(ctx)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "no session")
	}
	authSession, err := g.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no auth session")
	}

	gameInfo, err := g.gameService.GetGame(ctx.Request().Context(), authSession, values.GameID(gameID))
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	} else if err != nil {
		log.Printf("error: failed to get game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game")
	}

	resOwners := make([]string, 0, len(gameInfo.Owners))
	for _, ownerInfo := range gameInfo.Owners {
		resOwners = append(resOwners, string(ownerInfo.GetName()))
	}
	resMaintainers := make([]string, 0, len(gameInfo.Maintainers))
	for _, maintainerInfo := range gameInfo.Maintainers {
		resMaintainers = append(resMaintainers, string(maintainerInfo.GetName()))
	}

	resGenres := make([]string, 0, len(gameInfo.Genres))
	for _, genre := range gameInfo.Genres {
		resGenres = append(resGenres, string(genre.GetName()))
	}

	res := openapi.Game{
		Name:        string(gameInfo.Game.GetName()),
		Id:          uuid.UUID(gameInfo.Game.GetID()),
		Description: string(gameInfo.Game.GetDescription()),
		CreatedAt:   gameInfo.Game.GetCreatedAt(),
		Owners:      resOwners,
		Maintainers: &resMaintainers,
		Genres:      &resGenres,
	}
	return ctx.JSON(http.StatusOK, res)
}

// ゲームの情報の変更
// (PATCH /games/{gameID})
func (g *Game) PatchGame(ctx echo.Context, gameID openapi.GameIDInPath) error {
	req := openapi.PatchGameJSONRequestBody{}
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	gameName := values.NewGameName(req.Name)
	err = gameName.Validate()
	if errors.Is(err, values.ErrGameNameEmpty) {
		return echo.NewHTTPError(http.StatusBadRequest, "game name is empty")
	}
	if errors.Is(err, values.ErrGameNameTooLong) {
		return echo.NewHTTPError(http.StatusBadRequest, "game name is too long")
	}
	if err != nil {
		log.Printf("error: failed to validate game name: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to validate game name")
	}

	game, err := g.gameService.UpdateGame(
		ctx.Request().Context(),
		values.GameID(gameID),
		gameName,
		values.GameDescription(req.Description),
	)
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	} else if err != nil {
		log.Printf("error: failed to update game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update game")
	}

	res := openapi.GameInfo{
		Name:        string(game.GetName()),
		Id:          uuid.UUID(game.GetID()),
		Description: string(game.GetDescription()),
		CreatedAt:   game.GetCreatedAt(),
	}

	return ctx.JSON(http.StatusOK, res)
}
