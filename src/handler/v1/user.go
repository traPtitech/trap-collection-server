package v1

import "github.com/traPtitech/trap-collection-server/src/service"

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
