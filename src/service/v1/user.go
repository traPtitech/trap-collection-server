package v1

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/traPtitech/trap-collection-server/src/auth"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User struct {
	userUtils *UserUtils
}

func NewUser(userUtils *UserUtils) *User {
	return &User{
		userUtils: userUtils,
	}
}

func (u *User) GetMe(ctx context.Context, session *domain.OIDCSession) (*service.UserInfo, error) {
	return u.userUtils.getMe(ctx, session)
}

func (u *User) GetAllActiveUser(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error) {
	return u.userUtils.getAllActiveUser(ctx, session)
}

/*
	UserUtils
	traPメンバーの情報取得周り色々。
	TODO: 名前をもちょっとどうにかしたい。
*/
type UserUtils struct {
	userAuth  auth.User
	userCache cache.User
}

func NewUserUtils(userAuth auth.User, userCache cache.User) *UserUtils {
	return &UserUtils{
		userAuth:  userAuth,
		userCache: userCache,
	}
}

func (uu *UserUtils) getMe(ctx context.Context, session *domain.OIDCSession) (*service.UserInfo, error) {
	user, err := uu.userCache.GetMe(ctx, session.GetAccessToken())
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		// cacheからの取り出しに失敗してもauthからとって来れれば良いので、returnはしない
		log.Printf("error: failed to get user info: %v\n", err)
	}
	// cacheから取り出した場合はそれを返す
	if err == nil {
		return user, nil
	}

	user, err = uu.userAuth.GetMe(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	err = uu.userCache.SetMe(ctx, session, user)
	if err != nil {
		// cacheの設定に失敗してもreturnはしない
		log.Printf("error: failed to set user info: %v\n", err)
	}

	return user, nil
}

func (uu *UserUtils) getAllActiveUser(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error) {
	users, err := uu.userCache.GetAllActiveUsers(ctx)
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		// cacheからの取り出しに失敗してもauthからとって来れれば良いので、returnはしない
		log.Printf("error: failed to get user info: %v\n", err)
	}
	// cacheから取り出した場合はそれを返す
	if err == nil {
		return users, nil
	}

	users, err = uu.userAuth.GetAllActiveUsers(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	err = uu.userCache.SetAllActiveUsers(ctx, users)
	if err != nil {
		// cacheの設定に失敗してもreturnはしない
		log.Printf("error: failed to set user info: %v\n", err)
	}

	return users, nil
}
