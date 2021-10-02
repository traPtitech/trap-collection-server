package domain

import (
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

/*
	LauncherSession
	ランチャーのプロダクトキーでの認証後のセッションを表すドメイン。
	プロダクトキーでの認証から一定時間を過ぎると無効になり、
	再度認証が必要になる。
	有効期限の延長は不可能。
*/
type LauncherSession struct {
	id          values.LauncherSessionID
	accessToken values.LauncherSessionAccessToken
	expiresAt   time.Time
}

func NewLauncherSession(
	id values.LauncherSessionID,
	accessToken values.LauncherSessionAccessToken,
	expiresAt time.Time,
) *LauncherSession {
	return &LauncherSession{
		id:          id,
		accessToken: accessToken,
		expiresAt:   expiresAt,
	}
}

func (ls *LauncherSession) GetID() values.LauncherSessionID {
	return ls.id
}

func (ls *LauncherSession) GetAccessToken() values.LauncherSessionAccessToken {
	return ls.accessToken
}

func (ls *LauncherSession) GetExpiresAt() time.Time {
	return ls.expiresAt
}

// IsExpired 有効期限を過ぎていたらtrue
func (ls *LauncherSession) IsExpired() bool {
	return time.Now().After(ls.expiresAt)
}
