package repository

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type LauncherVersion interface {
	CreateLauncherVersion(context.Context, *domain.LauncherVersion) error
	GetLauncherVersions(context.Context) ([]*domain.LauncherVersion, error)
	GetLauncherVersion(context.Context, values.LauncherVersionID, LockType) (*domain.LauncherVersion, error)
	GetLauncherUsersByLauncherVersionID(context.Context, values.LauncherVersionID) ([]*domain.LauncherUser, error)
	GetLauncherVersionAndUserAndSessionByAccessToken(context.Context, values.LauncherSessionAccessToken) (*domain.LauncherVersion, *domain.LauncherUser, *domain.LauncherSession, error)
	AddGamesToLauncherVersion(context.Context, values.LauncherVersionID, []values.GameID) error
}
