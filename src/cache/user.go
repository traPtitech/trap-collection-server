package cache

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User interface {
	GetMe(ctx context.Context, accessToken values.OIDCAccessToken) (*service.UserInfo, error)
	SetMe(ctx context.Context, session *domain.OIDCSession, user *service.UserInfo) error
	// GetAllActiveUsers
	// deplicated
	// メソッド名にAllをつけない方が統一感があって良いので、GetActiveUsersに変更する
	// v1 API削除時に廃止する
	GetAllActiveUsers(ctx context.Context) ([]*service.UserInfo, error)
	// SetAllActiveUsers
	// deplicated
	// メソッド名にAllをつけない方が統一感があって良いので、SetActiveUsersに変更する
	// v1 API削除時に廃止する
	SetAllActiveUsers(ctx context.Context, users []*service.UserInfo) error
	// GetActiveUsers
	// traQの全アクティブユーザー(凍結されていないユーザー)のキャッシュ取得。
	// キャッシュ設定から1時間の間有効。
	// このため、traQでの凍結・凍結解除の反映までに最大1時間の遅延が発生する点に注意。
	GetActiveUsers(ctx context.Context) ([]*service.UserInfo, error)
	// SetActiveUsers
	// traQの全アクティブユーザー(凍結されていないユーザー)のキャッシュ設定。
	SetActiveUsers(ctx context.Context, users []*service.UserInfo) error
}
