package v1

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/handler/v1/openapi"
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

func (u *User) GetMe(c echo.Context) (*openapi.User, error) {
	session, err := getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	authSession, err := u.session.getAuthSession(session)
	if err != nil {
		// middlewareでログイン済みなことは確認しているので、ここではエラーになりえないはず
		log.Printf("error: failed to get auth session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	userInfo, err := u.userService.GetMe(c.Request().Context(), authSession)
	if err != nil {
		log.Printf("error: failed to get user info: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	return &openapi.User{
		Id:   uuid.UUID(userInfo.GetID()).String(),
		Name: string(userInfo.GetName()),
	}, nil
}

func (u *User) GetUsers(c echo.Context) ([]*openapi.User, error) {
	session, err := getSession(c)
	if err != nil {
		log.Printf("error: failed to get session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	authSession, err := u.session.getAuthSession(session)
	if err != nil {
		// middlewareでログイン済みなことは確認しているので、ここではエラーになりえないはず
		log.Printf("error: failed to get auth session: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	userInfos, err := u.userService.GetAllActiveUser(c.Request().Context(), authSession)
	if err != nil {
		log.Printf("error: failed to get user info: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	users := make([]*openapi.User, 0, len(userInfos))
	for _, userInfo := range userInfos {
		users = append(users, &openapi.User{
			Id:   uuid.UUID(userInfo.GetID()).String(),
			Name: string(userInfo.GetName()),
		})
	}

	return users, nil
}
