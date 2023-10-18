package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre interface {
	GetGenresByGameID(ctx context.Context, gameID values.GameID) ([]*domain.GameGenre, error)
	// RemoveGameGenre
	// ゲームジャンルを削除する。
	// IDに該当するゲームジャンルが存在しない場合は、ErrNoRecordDeletedを返す。
	RemoveGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error
	// GetGameGenresWithNames
	// ジャンルの名前の配列を指定してゲームジャンルを取得する。
	// 該当するゲームジャンルが存在しない場合は、ErrRecordNotFoundを返す。
	GetGameGenresWithNames(ctx context.Context, gameGenreNames []values.GameGenreName) ([]*domain.GameGenre, error)
	// SaveGameGenres
	// ゲームジャンルを作成する。
	// 名前が重複するゲームジャンルが1つでも存在するとき、ErrDuplicateUniqueKeyを返す。
	SaveGameGenres(ctx context.Context, gameGenres []*domain.GameGenre) error
	// RegisterGenresToGame
	// ゲームにゲームジャンルを登録する。
	RegisterGenresToGame(ctx context.Context, gameID values.GameID, gameGenres []values.GameGenreID) error
}
