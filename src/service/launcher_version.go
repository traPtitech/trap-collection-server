package service

//go:generate mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type LauncherVersion interface {
	CreateLauncherVersion(ctx context.Context, name values.LauncherVersionName, questionnaireURL values.LauncherVersionQuestionnaireURL) (*domain.LauncherVersion, error)
	GetLauncherVersions(ctx context.Context) ([]*domain.LauncherVersion, error)
	GetLauncherVersion(ctx context.Context, id values.LauncherVersionID) (*domain.LauncherVersion, []*domain.Game, error)
	AddGamesToLauncherVersion(ctx context.Context, id values.LauncherVersionID, gameIDs []values.GameID) (*domain.LauncherVersion, []*domain.Game, error)
	GetLauncherVersionCheckList(ctx context.Context, launcherVersionID values.LauncherVersionID, env *values.LauncherEnvironment) ([]*CheckListItem, error)
}

type CheckListItem struct {
	*domain.Game
	LatestVersion *domain.GameVersion
	/*
		LatestURL
		最新のゲームバージョンのURL
		nullableなことに注意!
	*/
	LatestURL *domain.GameURL
	/*
		LatestFile
		最新のゲームバージョンのファイル
		nullableなことに注意!
	*/
	LatestFile *domain.GameFile
	// LatestImage nullableでない
	LatestImage *domain.GameImage
	// LatestVideo nullableなことに注意!
	LatestVideo *domain.GameVideo
}
