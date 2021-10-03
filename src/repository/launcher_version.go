package repository

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type LauncherVersion interface {
	GetLauncherVersion(context.Context, values.LauncherVersionID) (*domain.LauncherVersion, error)
	GetLauncherUsersByLauncherVersionID(context.Context, values.LauncherVersionID) ([]*domain.LauncherUser, error)
	GetLauncherVersionAndUserAndSessionByAccessToken(context.Context, values.LauncherSessionAccessToken) (*domain.LauncherVersion, *domain.LauncherUser, *domain.LauncherSession, error)
}
