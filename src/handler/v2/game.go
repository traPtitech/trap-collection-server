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
	user        service.User
}

func NewGame(session *Session, gameService service.GameV2, user service.User) *Game {
	return &Game{
		session:     session,
		gameService: gameService,
		user:        user,
	}
}

// gameUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type gameUnimplemented interface {
	// ゲーム一覧の取得
	// (GET /games)
	GetGames(ctx echo.Context, params openapi.GetGamesParams) error
	// ゲームの追加
	// (POST /games)
	PostGame(ctx echo.Context) error
	// ゲームの削除
	// (DELETE /games/{gameID})
	DeleteGame(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲーム情報の取得
	// (GET /games/{gameID})
	GetGame(ctx echo.Context, gameID openapi.GameIDInPath) error
	// ゲームの情報の変更
	// (PATCH /games/{gameID})
	PatchGame(ctx echo.Context, gameID openapi.GameIDInPath) error
}

// ゲーム一覧の取得
// (GET /games)
func (g *Game) GetGames(ctx echo.Context, params openapi.GetGamesParams) error {
	isAll := bool(*params.All)

	var limit int
	var offset int
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	var games []*domain.Game
	var gameNumber int
	var err error
	if isAll {
		gameNumber, games, err = g.gameService.GetGames(ctx.Request().Context(), limit, offset)
		if err != nil {
			log.Printf("error: failed to get games: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
		}
	} else {
		session, err := g.session.get(ctx)
		if err != nil {
			log.Printf("error: failed to save session: %v\n", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
		}
		authSession, err := g.session.getAuthSession(session)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth session")
		}

		gameNumber, games, err = g.gameService.GetMyGames(ctx.Request().Context(), authSession, limit, offset)
		if err != nil {
			log.Printf("error: failed to get games: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
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
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}
	authSession, err := g.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth session")
	}

	req := openapi.NewGame{}
	err = ctx.Bind(req)

	owners := make([]values.TraPMemberName, 0, len(*req.Owners))
	for _, reqOwner := range *req.Owners {
		owners = append(owners, values.NewTrapMemberName(reqOwner))
	}

	maintainers := make([]values.TraPMemberName, 0, len(*req.Maintainers))
	for _, reqMaintainer := range *req.Maintainers {
		maintainers = append(maintainers, values.NewTrapMemberName(reqMaintainer))
	}

	gameInfo, err := g.gameService.CreateGame(
		ctx.Request().Context(),
		authSession, values.GameName(req.Name),
		values.GameDescription(req.Description),
		owners,
		maintainers)

	if errors.Is(err, service.ErrOverlapInOwners) {
		log.Printf("failed to add roles: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "failed to add owners")
	} else if errors.Is(err, service.ErrOverlapInMaintainers) {
		log.Printf("failed to add roles: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "failed to add maintainers")
	} else if errors.Is(err, service.ErrOverlapBetweenOwnersAndMaintainers) {
		log.Printf("failed to add roles: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "failed to add owners and maintainers")
	} else if err != nil {
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

	res := openapi.Game{
		Name:        string(gameInfo.Game.GetName()),
		Id:          uuid.UUID(gameInfo.Game.GetID()),
		Description: string(gameInfo.Game.GetDescription()),
		CreatedAt:   gameInfo.Game.GetCreatedAt(),
		Owners:      resOwners,
		Maintainers: &resOwners,
	}

	return ctx.JSON(http.StatusCreated, res)
}

// ゲームの削除
// (DELETE /games/{gameID})
func (g *Game) DeleteGame(ctx echo.Context, gameID openapi.GameIDInPath) error {
	session, err := g.session.get(ctx)
	if err != nil {
		log.Printf("error: failed to save session: %v\n", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}
	authSession, err := g.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth session")
	}

	userInfo, err := g.user.GetMe(ctx.Request().Context(), authSession)
	userName := userInfo.GetName()

	game, err := g.gameService.GetGame(ctx.Request().Context(), authSession, values.NewGameID())
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "Internal Server Error")
	} else if err != nil {
		log.Printf("failed to get game: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	ownersMap := make(map[values.TraPMemberName]struct{}, len(game.Owners))
	for _, ownerInfo := range game.Owners {
		ownersMap[ownerInfo.GetName()] = struct{}{}
	}
	if _, ok := ownersMap[userName]; !ok {
		return echo.NewHTTPError(http.StatusForbidden, "Internal Server Error")
	}

	err = g.gameService.DeleteGame(ctx.Request().Context(), values.GameID(gameID))
	if errors.Is(err, service.ErrNoGame) {
		//上のGetGameでやってるから起きなさそう
		return echo.NewHTTPError(http.StatusNotFound, "Internal Server Error")
	} else if err != nil {
		log.Printf("failed to get game: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	return ctx.NoContent(http.StatusOK)
}
