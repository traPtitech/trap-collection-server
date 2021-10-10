package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Middleware struct {
	session             *Session
	launcherAuthService service.LauncherAuth
	oidcService         service.OIDC
}

func NewMiddleware(session *Session, launcherAuthService service.LauncherAuth, oidcService service.OIDC) *Middleware {
	return &Middleware{
		session:             session,
		launcherAuthService: launcherAuthService,
		oidcService:         oidcService,
	}
}

func (m *Middleware) LauncherAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ok, err := m.checkLauncherAuth(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid launcher auth")
		}

		return next(c)
	}
}

func (m *Middleware) checkLauncherAuth(c echo.Context) (bool, error) {
	authorizationHeader := c.Request().Header.Get(echo.HeaderAuthorization)

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return false, fmt.Errorf("invalid authorization header: %s", authorizationHeader)
	}

	strAccessToken := strings.TrimPrefix(authorizationHeader, "Bearer ")
	accessToken := values.NewLauncherSessionAccessTokenFromString(strAccessToken)
	err := accessToken.Validate()
	if err != nil {
		return false, fmt.Errorf("invalid access token: %w", err)
	}

	launcherUser, launcherVersion, err := m.launcherAuthService.LauncherAuth(c.Request().Context(), accessToken)
	if errors.Is(err, service.ErrInvalidLauncherSessionAccessToken) {
		return false, fmt.Errorf("invalid access token: %w", err)
	}
	if errors.Is(err, service.ErrLauncherSessionAccessTokenExpired) {
		return false, fmt.Errorf("access token expired: %w", err)
	}
	if err != nil {
		return false, fmt.Errorf("failed to check launcher auth: %w", err)
	}

	c.Set(launcherUserKey, launcherUser)
	c.Set(launcherVersionKey, launcherVersion)

	return true, nil
}
