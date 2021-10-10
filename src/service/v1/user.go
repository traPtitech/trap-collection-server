package v1

import (
	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/cache"
)

type User struct {
	userAuth  auth.User
	userCache cache.User
}

func NewUser(userAuth auth.User, userCache cache.User) *User {
	return &User{
		userAuth:  userAuth,
		userCache: userCache,
	}
}
