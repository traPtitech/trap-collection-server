package cache

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User interface {
	GetMe(ctx context.Context, accessToken values.OIDCAccessToken) (*service.UserInfo, error)
	SetMe(ctx context.Context, session *domain.OIDCSession, user *service.UserInfo) error
	GetAllActiveUsers(ctx context.Context) ([]*service.UserInfo, error)
	SetAllActiveUsers(ctx context.Context, users []*service.UserInfo) error
}
