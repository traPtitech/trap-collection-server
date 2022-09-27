package auth

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User interface {
	GetMe(ctx context.Context, session *domain.OIDCSession) (*service.UserInfo, error)
	// GetAllActiveUsers
	// deplicated
	// メソッド名にAllをつけない方が統一感があって良いので、GetActiveUsersに変更する
	GetAllActiveUsers(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error)
	// GetActiveUsers
	// traQの全アクティブユーザー(凍結されていないユーザー)の取得。
	GetActiveUsers(ctx context.Context, session *domain.OIDCSession) ([]*service.UserInfo, error)
}
