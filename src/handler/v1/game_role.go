package v1

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameRole struct {
	featureWrite    bool
	session         *Session
	gameAuthService service.GameAuth
}

func NewGameRole(appConf config.App, session *Session, gameAuthService service.GameAuth) *GameRole {
	return &GameRole{
		featureWrite:    appConf.FeatureV1Write(),
		session:         session,
		gameAuthService: gameAuthService,
	}
}

func (gr *GameRole) PostMaintainer(c echo.Context, strGameID string, maintainers *openapi.Maintainers) error {
	if !gr.featureWrite {
		return echo.NewHTTPError(http.StatusForbidden, "write is disabled")
	}

	session, err := getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	authSession, err := gr.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return echo.NewHTTPError(http.StatusBadRequest, "no auth session")
	}
	if err != nil {
		log.Printf("error: failed to get auth session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth session")
	}

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	userIDs := make([]values.TraPMemberID, 0, len(maintainers.Maintainers))
	for _, maintainer := range maintainers.Maintainers {
		uuidUserID, err := uuid.Parse(maintainer)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid maintainer id")
		}

		userIDs = append(userIDs, values.NewTrapMemberID(uuidUserID))
	}

	err = gr.gameAuthService.AddGameCollaborators(c.Request().Context(), authSession, gameID, userIDs)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if errors.Is(err, service.ErrInvalidUserID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}
	if err != nil {
		log.Printf("error: failed to add game collaborators: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add game collaborators")
	}

	return nil
}

func (gr *GameRole) GetMaintainer(c echo.Context, strGameID string) ([]*openapi.Maintainer, error) {
	session, err := getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}

	authSession, err := gr.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no auth session")
	}
	if err != nil {
		log.Printf("error: failed to get auth session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get auth session")
	}

	uuidGameID, err := uuid.Parse(strGameID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}

	gameID := values.NewGameIDFromUUID(uuidGameID)

	gameManagers, err := gr.gameAuthService.GetGameManagers(c.Request().Context(), authSession, gameID)
	if errors.Is(err, service.ErrInvalidGameID) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
	}
	if err != nil {
		log.Printf("error: failed to get game managers: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get game managers")
	}

	maintainers := make([]*openapi.Maintainer, 0, len(gameManagers))
	for _, gameManager := range gameManagers {
		var role int32
		switch gameManager.Role {
		case values.GameManagementRoleAdministrator:
			role = 1
		case values.GameManagementRoleCollaborator:
			role = 0
		default:
			log.Printf("error: invalid game manager role: %d\n", gameManager.Role)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "invalid game manager role")
		}

		maintainers = append(maintainers, &openapi.Maintainer{
			Id:   uuid.UUID(gameManager.UserID).String(),
			Name: string(gameManager.UserName),
			Role: role,
		})
	}

	return maintainers, nil
}
