package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type Edition interface {
	// CreateEdition
	// エディションの作成。
	// ゲームバージョンが存在しない場合はErrInvalidGameVersionIDを返す。
	// ゲームバージョンが重複している場合はErrDuplicateGameVersionを返す。
	// ゲームが重複している場合はErrDuplicateGameを返す。
	CreateEdition(
		ctx context.Context,
		name values.LauncherVersionName,
		questionnaireURL types.Option[values.LauncherVersionQuestionnaireURL],
		gameVersionIDs []values.GameVersionID,
	) (*domain.LauncherVersion, error)
	// GetEditions
	// エディションの一覧の取得。
	GetEditions(ctx context.Context) ([]*domain.LauncherVersion, error)
	// GetEdition
	// エディションの取得。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	GetEdition(ctx context.Context, editionID values.LauncherVersionID) (*domain.LauncherVersion, error)
	// UpdateEdition
	// エディション情報の更新。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	// ゲームバージョンが重複している場合はErrDuplicateGameVersionを返す。
	// ゲームが重複している場合はErrDuplicateGameを返す。
	UpdateEdition(
		ctx context.Context,
		editionID values.LauncherVersionID,
		name values.LauncherVersionName,
		questionnaireURL types.Option[values.LauncherVersionQuestionnaireURL],
	) (*domain.LauncherVersion, error)
	// DeleteEdition
	// エディションの削除。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	DeleteEdition(ctx context.Context, editionID values.LauncherVersionID) error
	// UpdateEditionGameVersions
	// エディションに含まれるゲームバージョンの更新。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	// ゲームバージョンが存在しない場合はErrInvalidGameVersionIDを返す。
	UpdateEditionGameVersions(
		ctx context.Context,
		editionID values.LauncherVersionID,
		gameVersionIDs []values.GameVersionID,
	) ([]*GameVersionWithGame, error)
	// GetEditionGameVersions
	// エディションに含まれるゲームバージョンの取得。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	GetEditionGameVersions(ctx context.Context, editionID values.LauncherVersionID) ([]*GameVersionWithGame, error)
}

type GameVersionWithGame struct {
	GameVersion GameVersionInfo
	Game        *domain.Game
}
