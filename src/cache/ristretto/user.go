package ristretto

import (
	"context"
	"fmt"

	"github.com/dgraph-io/ristretto"
	"github.com/traPtitech/trap-collection-server/src/cache"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User struct {
	meCache     *ristretto.Cache
	activeUsers *ristretto.Cache
}

func NewUser() (*User, error) {
	meCache, err := ristretto.NewCache(&ristretto.Config{
		/*
			アクセス頻度を保持する要素の数。
			一般的には最大で格納される要素数の10倍程度が良いらしいが、
			最大でtraP部員数しか格納されないことを考えて500を設定する。
		*/
		NumCounters: 500,
		/*
			キャッシュの最大サイズ。
			あまり大きくしすぎるとメモリが足りなくなるので注意!
			*UserInfo1つあたり8Byteなので、8*500=20kB<2**15に設定する。
		*/
		MaxCost:     1 << 15,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create meCache: %v", err)
	}

	activeUsers, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10,
		MaxCost:     64,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create activeUsers: %v", err)
	}

	return &User{
		meCache:     meCache,
		activeUsers: activeUsers,
	}, nil
}

func (u *User) GetMe(ctx context.Context, accessToken values.OIDCAccessToken) (*service.UserInfo, error) {
	iUser, ok := u.meCache.Get(string(accessToken))
	if !ok {
		return nil, cache.ErrCacheMiss
	}

	user, ok := iUser.(*service.UserInfo)
	if !ok {
		return nil, fmt.Errorf("failed to cast meCache: %v", iUser)
	}

	return user, nil
}
