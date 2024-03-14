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
	session, err := g.session.get(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no session")
	}
	authSession, _ := g.session.getAuthSession(session)
	// authSessionが取得出来なくても、普通のユーザーとしてゲーム一覧を取得するため、エラーは返さない。
	var visibilities []values.GameVisibility
	if authSession == nil {
		visibilities = []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited}
	} else {
		visibilities = []values.GameVisibility{values.GameVisibilityTypePublic, values.GameVisibilityTypeLimited, values.GameVisibilityTypePrivate}
	}

	var isAll bool
	if params.All != nil && authSession != nil {
		isAll = *params.All
	} else {
		isAll = true
	}

	limit, offset := 0, 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	var sortType service.GamesSortType
	if params.Sort != nil {
		switch *params.Sort {
		case openapi.CreatedAt:
			sortType = service.GamesSortTypeCreatedAt
		case openapi.LatestVersion:
			sortType = service.GamesSortTypeLatestVersion
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid sort type")
		}
	} else {
		sortType = service.GamesSortTypeCreatedAt
	}

	var gameName string
	if params.Name != nil {
		gameName = *params.Name
	}

	var gameGenreIDs []values.GameGenreID
	if params.Genre != nil {
		gameGenreIDs = make([]values.GameGenreID, 0, len(*params.Genre))
		for i := range *params.Genre {
			gameGenreIDs = append(gameGenreIDs, values.GameGenreIDFromUUID((*params.Genre)[i]))
		}
	}

	var gameWithGenres []*domain.GameWithGenres
	var gameNumber int
	if isAll {
		gameNumber, gameWithGenres, err = g.gameService.GetGames(ctx.Request().Context(), limit, offset, sortType, visibilities, gameGenreIDs, gameName)
		if err != nil {
			log.Printf("error: failed to get games: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get games")
		}
	} else {
		gameNumber, gameWithGenres, err = g.gameService.GetMyGames(ctx.Request().Context(), authSession, limit, offset, sortType, visibilities, gameGenreIDs, gameName)
		if err != nil {
			log.Printf("error: failed to get games: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get my games")
		}
	}

	responseGames := make([]openapi.GameInfoWithGenres, 0, len(gameWithGenres))
	for i := range gameWithGenres {
		var visibility openapi.GameVisibility
		switch gameWithGenres[i].GetGame().GetVisibility() {
		case values.GameVisibilityTypePublic:
			visibility = openapi.Public
		case values.GameVisibilityTypeLimited:
			visibility = openapi.Limited
		case values.GameVisibilityTypePrivate:
			visibility = openapi.Private
		default:
			log.Printf("error: failed to get game visibility: %v\n", gameWithGenres[i].GetGame().GetVisibility())
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get games")
		}

		genreNames := make([]openapi.GameGenreName, 0, len(gameWithGenres[i].GetGenres()))
		for _, genre := range gameWithGenres[i].GetGenres() {
			genreNames = append(genreNames, openapi.GameGenreName(genre.GetName()))
		}

		responseGame := openapi.GameInfoWithGenres{
			Name:        string(gameWithGenres[i].GetGame().GetName()),
			Id:          uuid.UUID(gameWithGenres[i].GetGame().GetID()),
			Description: string(gameWithGenres[i].GetGame().GetDescription()),
			Visibility:  visibility,
			CreatedAt:   gameWithGenres[i].GetGame().GetCreatedAt(),
			Genres:      &genreNames,
		}
		responseGames = append(responseGames, responseGame)
	}

	res := openapi.GetGamesResponse{
		Games: responseGames,
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
		// 部員以外でも、管理者情報以外は取得できるようにするので、エラーは返さない。
		authSession = nil
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

	var visibility openapi.GameVisibility
	switch gameInfo.Game.GetVisibility() {
	case values.GameVisibilityTypePublic:
		visibility = openapi.Public
	case values.GameVisibilityTypeLimited:
		visibility = openapi.Limited
	case values.GameVisibilityTypePrivate:
		visibility = openapi.Private
	default:
		log.Printf("error: failed to get game visibility: %v\n", gameInfo.Game.GetVisibility())
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game visibility")
	}

	res := openapi.Game{
		Name:        string(gameInfo.Game.GetName()),
		Id:          uuid.UUID(gameInfo.Game.GetID()),
		Description: string(gameInfo.Game.GetDescription()),
		CreatedAt:   gameInfo.Game.GetCreatedAt(),
		Owners:      resOwners,
		Maintainers: &resMaintainers,
		Genres:      &resGenres,
		Visibility:  visibility,
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
