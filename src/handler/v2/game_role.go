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

type GameRole struct {
	gameRoleService service.GameRoleV2
	gameService     service.GameV2
	session         *Session
}

func NewGameRole(gameRoleService service.GameRoleV2, gameService service.GameV2, session *Session) *GameRole {
	return &GameRole{
		gameRoleService: gameRoleService,
		gameService:     gameService,
		session:         session,
	}
}

func (gameRole *GameRole) PatchGameRole(ctx echo.Context, gameID openapi.GameIDInPath) error {
	session, err := gameRole.session.get(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no session")
	}
	authSession, err := gameRole.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no auth session")
	}
	req := &openapi.PatchGameRoleJSONRequestBody{}
	err = ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	userID := values.NewTrapMemberID(req.Id)
	var roleType values.GameManagementRole
	switch *req.Type {
	case "owner":
		roleType = values.GameManagementRoleAdministrator
	case "maintainer":
		roleType = values.GameManagementRoleCollaborator
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "role type is invalid")
	}

	err = gameRole.gameRoleService.EditGameManagementRole(ctx.Request().Context(), authSession, values.GameID(gameID), userID, roleType)
	if errors.Is(err, service.ErrNoGameManagementRoleUpdated) {
		return echo.NewHTTPError(http.StatusBadRequest, "there is no change")
	}
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusNotFound, "no game")
	}
	if errors.Is(err, service.ErrInvalidUserID) {
		return echo.NewHTTPError(http.StatusBadRequest, "userID is invalid or no user")
	}
	if errors.Is(err, service.ErrCannotEditOwners) {
		return echo.NewHTTPError(http.StatusBadRequest, "you cannot change the user role beause there is only 1 owner")
	}
	if err != nil {
		log.Printf("error: failed to edit game management role: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to edit game management role")
	}

	newGameInfo, err := gameRole.gameService.GetGame(ctx.Request().Context(), authSession, values.GameID(gameID))
	if errors.Is(err, service.ErrNoGame) {
		//上でおんなじことやってるけど一応
		return echo.NewHTTPError(http.StatusNotFound, "no game")
	}
	if err != nil {
		log.Printf("error: failed to get game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game")
	}

	resOwners := make([]string, 0, len(newGameInfo.Owners))
	for _, owner := range newGameInfo.Owners {
		resOwners = append(resOwners, string(owner.GetName()))
	}

	resMaintainers := make([]string, 0, len(newGameInfo.Maintainers))
	for _, maintainer := range newGameInfo.Maintainers {
		resMaintainers = append(resMaintainers, string(maintainer.GetName()))
	}

	var visibility openapi.GameVisibility
	visibility, err = convertGameVisibility(newGameInfo.Game.GetVisibility())
	if err != nil {
		log.Printf("error: failed to convert game visibility: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to convert game visibility")
	}

	genres := make([]openapi.GameGenreName, 0, len(newGameInfo.Genres))
	for _, genre := range newGameInfo.Genres {
		genres = append(genres, openapi.GameGenreName(genre.GetName()))
	}

	resGame := openapi.Game{
		Id:          uuid.UUID(newGameInfo.Game.GetID()),
		Name:        string(newGameInfo.Game.GetName()),
		Description: string(newGameInfo.Game.GetDescription()),
		CreatedAt:   newGameInfo.Game.GetCreatedAt(),
		Visibility:  visibility,
		Owners:      resOwners,
		Maintainers: &resMaintainers,
		Genres:      &genres,
	}

	return ctx.JSON(http.StatusOK, resGame)
}

// ゲームの管理権限の削除
// (DELETE /games/{gameID}/roles/{userID})
func (gameRole *GameRole) DeleteGameRole(ctx echo.Context, gameID openapi.GameIDInPath, userID openapi.UserIDInPath) error {
	session, err := gameRole.session.get(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no session")
	}
	authSession, err := gameRole.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no auth session")
	}

	err = gameRole.gameRoleService.RemoveGameManagementRole(ctx.Request().Context(), values.GameID(gameID), values.TraPMemberID(userID))
	if errors.Is(err, service.ErrInvalidRole) {
		return echo.NewHTTPError(http.StatusNotFound, "the user does not has any role")
	}
	if errors.Is(err, service.ErrCannotDeleteOwner) {
		return echo.NewHTTPError(http.StatusBadRequest, "you cannot delete owner because there is only 1 owner")
	}
	if errors.Is(err, service.ErrNoGame) {
		return echo.NewHTTPError(http.StatusBadRequest, "no game")
	}
	if err != nil {
		log.Printf("error: failed to remove game management role: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove game management role")
	}

	newGameInfo, err := gameRole.gameService.GetGame(ctx.Request().Context(), authSession, values.GameID(gameID))
	if errors.Is(err, service.ErrNoGame) {
		//上でおんなじことやってるけど一応
		return echo.NewHTTPError(http.StatusNotFound, "no game")
	}
	if err != nil {
		log.Printf("error: failed to get game: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game")
	}

	resOwners := make([]string, 0, len(newGameInfo.Owners))
	for _, owner := range newGameInfo.Owners {
		resOwners = append(resOwners, string(owner.GetName()))
	}

	resMaintainers := make([]string, 0, len(newGameInfo.Maintainers))
	for _, maintainer := range newGameInfo.Maintainers {
		resMaintainers = append(resMaintainers, string(maintainer.GetName()))
	}

	resGame := openapi.Game{
		Id:          uuid.UUID(newGameInfo.Game.GetID()),
		Name:        string(newGameInfo.Game.GetName()),
		Description: string(newGameInfo.Game.GetDescription()),
		CreatedAt:   newGameInfo.Game.GetCreatedAt(),
		Owners:      resOwners,
		Maintainers: &resMaintainers,
	}

	return ctx.JSON(http.StatusOK, resGame)
}
