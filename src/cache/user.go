package cache

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type User interface {
	GetMe(ctx context.Context, session *domain.OIDCSession) (*service.UserInfo, error)
	GetAllActiveUsers(ctx context.Context) ([]*service.UserInfo, error)
}
