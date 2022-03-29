package repository

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type LauncherUser interface {
	CreateLauncherUsers(context.Context, values.LauncherVersionID, []*domain.LauncherUser) ([]*domain.LauncherUser, error)
	DeleteLauncherUser(context.Context, values.LauncherUserID) error
	GetLauncherUserByProductKey(context.Context, values.LauncherUserProductKey) (*domain.LauncherUser, error)
}
