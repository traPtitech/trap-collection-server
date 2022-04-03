package repository

//go:generate go run github.com/golang/mock/mockgen@latest -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Game interface {
	SaveGame(ctx context.Context, game *domain.Game) error
	UpdateGame(ctx context.Context, game *domain.Game) error
	RemoveGame(ctx context.Context, gameID values.GameID) error
	GetGame(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.Game, error)
	GetGames(ctx context.Context) ([]*domain.Game, error)
	GetGamesByUser(ctx context.Context, userID values.TraPMemberID) ([]*domain.Game, error)
	GetGamesByIDs(ctx context.Context, gameIDs []values.GameID, lockType LockType) ([]*domain.Game, error)
	GetGamesByLauncherVersion(ctx context.Context, launcherVersionID values.LauncherVersionID) ([]*domain.Game, error)
	GetGameInfosByLauncherVersion(ctx context.Context, launcherVersionID values.LauncherVersionID, fileTypes []values.GameFileType) ([]*GameInfo, error)
}

type GameInfo struct {
	*domain.Game
	LatestVersion *domain.GameVersion
	/*
		LatestURL
		最新のゲームバージョンのURL
		nullableなことに注意!
	*/
	LatestURL *domain.GameURL
	/*
		LatestFiles
		最新のゲームバージョンのファイル
	*/
	LatestFiles []*domain.GameFile
	// LatestImage nullableなことに注意!
	LatestImage *domain.GameImage
	// LatestVideo nullableなことに注意!
	LatestVideo *domain.GameVideo
}
