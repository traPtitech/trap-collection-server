package v2

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi "github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User struct {
	session     *Session
	userService service.User
}

func NewUser(session *Session, userService service.User) *User {
	return &User{
		session:     session,
		userService: userService,
	}
}

func (u *User) GetMe(c echo.Context) error {
	session, err := u.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	authSession, err := u.session.getAuthSession(session)
	if err != nil {
		// middlewareでログイン済みなことは確認しているので、ここではエラーになりえないはず
		log.Printf("error: failed to get auth session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	userInfo, err := u.userService.GetMe(c.Request().Context(), authSession)
	if err != nil {
		log.Printf("error: failed to get user info: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, openapi.User{
		Id:   uuid.UUID(userInfo.GetID()),
		Name: string(userInfo.GetName()),
	})
}

func (u *User) GetUsers(c echo.Context) error {
	session, err := u.session.get(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	authSession, err := u.session.getAuthSession(session)
	if err != nil {
		// middlewareでログイン済みなことは確認しているので、ここではエラーになりえないはず
		log.Printf("error: failed to get auth session: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	userInfos, err := u.userService.GetAllActiveUser(c.Request().Context(), authSession)
	if err != nil {
		log.Printf("error: failed to get user info: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	users := make([]*openapi.User, 0, len(userInfos))
	for _, userInfo := range userInfos {
		users = append(users, &openapi.User{
			Id:   uuid.UUID(userInfo.GetID()),
			Name: string(userInfo.GetName()),
		})
	}

	return c.JSON(http.StatusOK, users)
}
