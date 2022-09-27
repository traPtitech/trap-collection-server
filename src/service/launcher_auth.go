package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type LauncherAuth interface {
	CreateLauncherUser(ctx context.Context, launcherVersionID values.LauncherVersionID, userNum int) ([]*domain.LauncherUser, error)
	GetLauncherUsers(ctx context.Context, launcherVersionID values.LauncherVersionID) ([]*domain.LauncherUser, error)
	RevokeProductKey(ctx context.Context, user values.LauncherUserID) error
	LoginLauncher(ctx context.Context, productKey values.LauncherUserProductKey) (*domain.LauncherSession, error)
	LauncherAuth(ctx context.Context, accessToken values.LauncherSessionAccessToken) (*domain.LauncherUser, *domain.LauncherVersion, error)
}
