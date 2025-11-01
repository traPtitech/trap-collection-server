package service

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// mockgenがgenericsに対応するまでの暫定対応
// interface内にgenericsの構文が出なければ問題ないのを利用している
// ref: https://github.com/golang/mock/pull/640
// TODO: mockgenのv1.7.0がリリースされ次第削除する
type (
	OptionQuestionnaireURL = types.Option[values.EditionQuestionnaireURL]
)

type Edition interface {
	// CreateEdition
	// エディションの作成。
	// ゲームバージョンが存在しない場合はErrInvalidGameVersionIDを返す。
	// ゲームバージョンが重複している場合はErrDuplicateGameVersionを返す。
	// ゲームが重複している場合はErrDuplicateGameを返す。
	CreateEdition(
		ctx context.Context,
		name values.EditionName,
		questionnaireURL OptionQuestionnaireURL,
		gameVersionIDs []values.GameVersionID,
	) (*domain.Edition, error)
	// GetEditions
	// エディションの一覧の取得。
	GetEditions(ctx context.Context) ([]*domain.Edition, error)
	// GetEdition
	// エディションの取得。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	GetEdition(ctx context.Context, editionID values.EditionID) (*domain.Edition, error)
	// UpdateEdition
	// エディション情報の更新。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	// ゲームバージョンが重複している場合はErrDuplicateGameVersionを返す。
	// ゲームが重複している場合はErrDuplicateGameを返す。
	UpdateEdition(
		ctx context.Context,
		editionID values.EditionID,
		name values.EditionName,
		questionnaireURL OptionQuestionnaireURL,
	) (*domain.Edition, error)
	// DeleteEdition
	// エディションの削除。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	DeleteEdition(ctx context.Context, editionID values.EditionID) error
	// UpdateEditionGameVersions
	// エディションに含まれるゲームバージョンの更新。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	// ゲームバージョンが存在しない場合はErrInvalidGameVersionIDを返す。
	UpdateEditionGameVersions(
		ctx context.Context,
		editionID values.EditionID,
		gameVersionIDs []values.GameVersionID,
	) ([]*GameVersionWithGame, error)
	// GetEditionGameVersions
	// エディションに含まれるゲームバージョンの取得。
	// エディションが存在しない場合はErrInvalidEditionIDを返す。
	GetEditionGameVersions(ctx context.Context, editionID values.EditionID) ([]*GameVersionWithGame, error)
}

type GameVersionWithGame struct {
	GameVersion GameVersionInfo
	Game        *domain.Game
}
