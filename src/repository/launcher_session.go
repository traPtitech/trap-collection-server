package repository

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type LauncherSession interface {
	CreateLauncherSession(ctx context.Context, launcherUserID values.LauncherUserID, launcherSession *domain.LauncherSession) (*domain.LauncherSession, error)
}
