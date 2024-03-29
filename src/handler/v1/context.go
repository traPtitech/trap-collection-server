package v1

import (
	"errors"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
)

const (
	launcherUserKey    = "launcherUser"
	launcherVersionKey = "launcherVersion"
	sessionContextKey  = "session"
)

func getLauncherUser(c echo.Context) (*domain.LauncherUser, error) {
	iLauncherUser := c.Get(launcherUserKey)
	if iLauncherUser == nil {
		return nil, errors.New("launcher user is not set")
	}

	launcherUser, ok := iLauncherUser.(*domain.LauncherUser)
	if !ok {
		return nil, errors.New("invalid launcher user")
	}

	return launcherUser, nil
}

func getLauncherVersion(c echo.Context) (*domain.LauncherVersion, error) {
	iLauncherVersion := c.Get(launcherVersionKey)
	if iLauncherVersion == nil {
		return nil, errors.New("launcher version is not set")
	}

	launcherVersion, ok := iLauncherVersion.(*domain.LauncherVersion)
	if !ok {
		return nil, errors.New("invalid launcher version")
	}

	return launcherVersion, nil
}

func getSession(c echo.Context) (*sessions.Session, error) {
	iSession := c.Get(sessionContextKey)
	if iSession == nil {
		return nil, errors.New("session is not set")
	}

	session, ok := iSession.(*sessions.Session)
	if !ok {
		return nil, errors.New("invalid session")
	}

	return session, nil
}
