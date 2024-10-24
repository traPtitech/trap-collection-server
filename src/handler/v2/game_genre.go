package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameGenre struct {
	gameGenreService service.GameGenre
	game             service.GameV2
	session          *Session
}

func NewGameGenre(gameGenreService service.GameGenre, game service.GameV2, session *Session) *GameGenre {
	return &GameGenre{
		gameGenreService: gameGenreService,
		game:             game,
		session:          session,
	}
}

// ジャンルの削除
// (DELETE /genres/{gameGenreID})
func (gameGenre *GameGenre) DeleteGameGenre(c echo.Context, gameGenreID openapi.GameGenreIDInPath) error {
	err := gameGenre.gameGenreService.DeleteGameGenre(c.Request().Context(), values.GameGenreIDFromUUID(gameGenreID))
	if errors.Is(err, service.ErrNoGameGenre) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid game genre ID")
	}
	if err != nil {
		log.Printf("error: failed to delete game genre: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete game genre")
	}

	return c.NoContent(http.StatusOK)
}

// 全てのジャンルの取得
// (GET /genres)
func (gameGenre *GameGenre) GetGameGenres(ctx echo.Context) error {
	session, err := gameGenre.session.get(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	// ログインしているかどうかだけ知ればいいので、auth sessionは捨てる
	_, err = gameGenre.session.getAuthSession(session)
	isLoginUser := (err == nil)

	gameGenreInfos, err := gameGenre.gameGenreService.GetGameGenres(ctx.Request().Context(), isLoginUser)
	if err != nil {
		log.Printf("error: failed to get game genres: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game genres")
	}

	gameGenresResponse := make([]openapi.GameGenre, len(gameGenreInfos))
	for i := range gameGenreInfos {
		gameGenresResponse[i] = openapi.GameGenre{
			Id:        uuid.UUID(gameGenreInfos[i].GetID()),
			Genre:     string(gameGenreInfos[i].GetName()),
			Num:       gameGenreInfos[i].Num,
			CreatedAt: gameGenreInfos[i].GetCreatedAt(),
		}
	}

	return ctx.JSON(http.StatusOK, gameGenresResponse)
}

// ゲームのジャンル編集
// (PUT /games/{gameID}/genres)
func (gameGenre *GameGenre) PutGameGenres(c echo.Context, gameID openapi.GameIDInPath) error {
	session, err := gameGenre.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}
	authSession, err := gameGenre.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var reqBody openapi.PutGameGenresJSONRequestBody
	if err := c.Bind(&reqBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	var gameGenreNames []values.GameGenreName
	if reqBody.Genres != nil {
		gameGenreNames = make([]values.GameGenreName, len(*reqBody.Genres))
		for i, genre := range *reqBody.Genres {
			gameGenreNames[i] = values.GameGenreName(genre)
		}
	} else {
		gameGenreNames = []values.GameGenreName{}
	}

	for i := range gameGenreNames {
		if gameGenreNames[i].Validate() != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid genre name")
		}
	}

	err = gameGenre.gameGenreService.UpdateGameGenres(c.Request().Context(), values.NewGameIDFromUUID(gameID), gameGenreNames)
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "invalid game ID")
	}
	if errors.Is(err, service.ErrDuplicateGameGenre) {
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate game genre")
	}
	if err != nil {
		log.Printf("error: failed to update game genres: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update game genres")
	}

	gameInfo, err := gameGenre.game.GetGame(c.Request().Context(), authSession, values.NewGameIDFromUUID(gameID))
	if err != nil {
		log.Printf("error: failed to get game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game")
	}

	gameGenresResponse := make([]openapi.GameGenreName, len(gameInfo.Genres))
	for i := range gameInfo.Genres {
		gameGenresResponse[i] = openapi.GameGenreName(gameInfo.Genres[i].GetName())
	}

	var visibility openapi.GameVisibility
	switch gameInfo.Game.GetVisibility() {
	case values.GameVisibilityTypePublic:
		visibility = openapi.Public
	case values.GameVisibilityTypeLimited:
		visibility = openapi.Limited
	case values.GameVisibilityTypePrivate:
		visibility = openapi.Private
	}

	owners := make([]openapi.UserName, len(gameInfo.Owners))
	for i := range gameInfo.Owners {
		owners[i] = openapi.UserName(gameInfo.Owners[i].GetName())
	}

	maintainers := make([]openapi.UserName, len(gameInfo.Maintainers))
	for i := range gameInfo.Maintainers {
		maintainers[i] = openapi.UserName(gameInfo.Maintainers[i].GetName())
	}

	res := openapi.Game{
		Id:          uuid.UUID(gameInfo.Game.GetID()),
		Name:        openapi.GameName(gameInfo.Game.GetName()),
		Description: openapi.GameDescription(gameInfo.Game.GetDescription()),
		Visibility:  visibility,
		Owners:      owners,
		CreatedAt:   gameInfo.Game.GetCreatedAt(),
	}

	if len(gameGenresResponse) > 0 {
		res.Genres = &gameGenresResponse
	}

	if len(maintainers) > 0 {
		res.Maintainers = &maintainers
	}

	return c.JSON(http.StatusOK, res)
}

// ジャンル情報の変更
// (PATCH /genres/{gameGenreID})
func (gameGenre *GameGenre) PatchGameGenre(c echo.Context, gameGenreID openapi.GameGenreIDInPath) error {
	var reqBody openapi.PatchGameGenreJSONRequestBody
	if err := c.Bind(&reqBody); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	gameGenreName := values.NewGameGenreName(reqBody.Genre)
	if err := gameGenreName.Validate(); err != nil {
		if errors.Is(err, values.ErrGameGenreNameEmpty) {
			return echo.NewHTTPError(http.StatusBadRequest, "genre name must not be empty")
		}
		if errors.Is(err, values.ErrGameGenreNameTooLong) {
			return echo.NewHTTPError(http.StatusBadRequest, "genre name is too long")
		}
		log.Printf("failed to validate genre name: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to validate genre name")
	}

	ctx := c.Request().Context()

	err := gameGenre.gameGenreService.UpdateGameGenre(ctx, values.GameGenreIDFromUUID(gameGenreID), gameGenreName)
	if errors.Is(err, service.ErrNoGameGenre) {
		return echo.NewHTTPError(http.StatusNotFound, "game genre not found")
	}
	if errors.Is(err, service.ErrDuplicateGameGenreName) {
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate genre name")
	}
	if errors.Is(err, service.ErrNoGameGenreUpdated) {
		return echo.NewHTTPError(http.StatusBadRequest, "no game genre updated")
	}
	if err != nil {
		log.Printf("error: failed to update game genre: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update game genre")
	}

	gameGenreInfo, err := gameGenre.gameGenreService.GetGameGenre(ctx, values.GameGenreIDFromUUID(gameGenreID))
	if errors.Is(err, service.ErrNoGameGenre) {
		return echo.NewHTTPError(http.StatusNotFound, "game genre not found")
	}
	if err != nil {
		log.Printf("error: failed to get game genre: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game genre")
	}

	res := openapi.GameGenre{
		Id:        uuid.UUID(gameGenreInfo.GetID()),
		Genre:     string(gameGenreInfo.GetName()),
		Num:       gameGenreInfo.Num,
		CreatedAt: gameGenreInfo.GetCreatedAt(),
	}

	return c.JSON(http.StatusOK, res)
}
