package v1

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Middleware struct {
	session             *Session
	launcherAuthService service.LauncherAuth
	gameAuthService     service.GameAuth
	oidcService         service.OIDC
}

func NewMiddleware(
	session *Session,
	launcherAuthService service.LauncherAuth,
	gameAuthService service.GameAuth,
	oidcService service.OIDC,
) *Middleware {
	return &Middleware{
		session:             session,
		launcherAuthService: launcherAuthService,
		gameAuthService:     gameAuthService,
		oidcService:         oidcService,
	}
}

func (m *Middleware) TrapMemberAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ok, message, err := m.checkTrapMemberAuth(c)
		if err != nil {
			log.Printf("error: failed to check trap member auth: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, message)
		}

		return next(c)
	}
}

func (m *Middleware) LauncherAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ok, message, err := m.checkLauncherAuth(c)
		if err != nil {
			log.Printf("error: failed to check launcher auth: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, message)
		}

		return next(c)
	}
}

func (m *Middleware) checkTrapMemberAuth(c echo.Context) (bool, string, error) {
	session, err := m.session.getSession(c)
	if err != nil {
		return false, "", fmt.Errorf("failed to get session: %w", err)
	}

	authSession, err := m.session.getAuthSession(session)
	if errors.Is(err, ErrNoValue) {
		return false, "no access token", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to get auth session: %w", err)
	}

	err = m.oidcService.TraPAuth(c.Request().Context(), authSession)
	if errors.Is(err, service.ErrOIDCSessionExpired) {
		return false, "access token is expired", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to check traP auth: %w", err)
	}

	return true, "", nil
}

func (m *Middleware) BothAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ok, launcherAuthMessage, err := m.checkLauncherAuth(c)
		if err != nil {
			log.Printf("error: failed to check launcher auth: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if ok {
			return next(c)
		}

		ok, traPAuthMessage, err := m.checkTrapMemberAuth(c)
		if err != nil {
			log.Printf("error: failed to check trap member auth: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Launcher Auth: %s\ntraP Auth: %s", launcherAuthMessage, traPAuthMessage))
		}

		return next(c)
	}
}

func (m *Middleware) checkLauncherAuth(c echo.Context) (bool, string, error) {
	authorizationHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return false, fmt.Sprintf("invalid authorization header: %s", authorizationHeader), nil
	}

	strAccessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	accessToken := values.NewLauncherSessionAccessTokenFromString(strAccessToken)
	err := accessToken.Validate()
	if err != nil {
		return false, fmt.Sprintf("invalid access token: %s", accessToken), nil
	}

	launcherUser, launcherVersion, err := m.launcherAuthService.LauncherAuth(c.Request().Context(), accessToken)
	if errors.Is(err, service.ErrInvalidLauncherSessionAccessToken) {
		return false, "invalid access token", nil
	}
	if errors.Is(err, service.ErrLauncherSessionAccessTokenExpired) {
		return false, "access token expired", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to check launcher auth: %w", err)
	}

	c.Set(launcherUserKey, launcherUser)
	c.Set(launcherVersionKey, launcherVersion)

	return true, "", nil
}

func (m *Middleware) GameMaintainerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := getSession(c)
		if err != nil {
			log.Printf("error: failed to get session: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		authSession, err := m.session.getAuthSession(session)
		if err != nil {
			// TrapMemberAuthMiddlewareでErrNoValueなどは弾かれているはずなので、ここでエラーは起きないはず
			log.Printf("error: failed to get auth session: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		strGameID := c.Param("gameID")
		uuidGameID, err := uuid.Parse(strGameID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
		}
		gameID := values.NewGameIDFromUUID(uuidGameID)

		err = m.gameAuthService.UpdateGameAuth(c.Request().Context(), authSession, gameID)
		if errors.Is(err, service.ErrInvalidGameID) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
		}
		if errors.Is(err, service.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		if err != nil {
			log.Printf("error: failed to update game auth: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return next(c)
	}
}

func (m *Middleware) GameOwnerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := getSession(c)
		if err != nil {
			log.Printf("error: failed to get session: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		authSession, err := m.session.getAuthSession(session)
		if err != nil {
			// TrapMemberAuthMiddlewareでErrNoValueなどは弾かれているはずなので、ここでエラーは起きないはず
			log.Printf("error: failed to get auth session: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		strGameID := c.Param("gameID")
		uuidGameID, err := uuid.Parse(strGameID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
		}
		gameID := values.NewGameIDFromUUID(uuidGameID)

		err = m.gameAuthService.UpdateGameManagementRoleAuth(c.Request().Context(), authSession, gameID)
		if errors.Is(err, service.ErrInvalidGameID) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid game id")
		}
		if errors.Is(err, service.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "forbidden")
		}
		if err != nil {
			log.Printf("error: failed to update game auth: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return next(c)
	}
}
