package repository

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
)

type LauncherSession interface {
	CreateLauncherSession(ctx context.Context, launcherSession *domain.LauncherSession) (*domain.LauncherSession, error)
}
