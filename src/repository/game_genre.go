package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre interface {
	// 全てのジャンルと、そのジャンルに含まれるゲームの数を、ジャンル作成日時の降順で返す。
	GetGameGenres(ctx context.Context, visibilities []values.GameVisibility) ([]*GameGenreInfo, error)
	// ゲームのIDからそのゲームのジャンルを取得する。
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
	// 名前が重複するゲームジャンルが1つでも存在するとき、ErrDuplicatedUniqueKeyを返す。
	SaveGameGenres(ctx context.Context, gameGenres []*domain.GameGenre) error
	// RegisterGenresToGame
	// ゲームにゲームジャンルを登録する。
	// ゲームが存在しない場合は、ErrRecordNotFoundを返す。
	// ゲームジャンルが存在しない場合は、ErrIncludeInvalidArgsを返す。
	// ゲームジャンルを追加するのではなく、置き換える。
	RegisterGenresToGame(ctx context.Context, gameID values.GameID, gameGenres []values.GameGenreID) error
	// UpdateGameGenre
	// ゲームジャンルの情報を更新する。
	// ゲームジャンルが存在しない場合は、ErrRecordNotFoundを返す。
	// ゲームジャンルの名前が重複する場合は、ErrDuplicatedUniqueKeyを返す。
	UpdateGameGenre(ctx context.Context, gameGenreID values.GameGenreID, gameGenre *domain.GameGenre) error
	// GetGameGenre
	// ゲームジャンルのIDからゲームジャンルを取得する。
	// ゲームジャンルが存在しない場合は、ErrRecordNotFoundを返す。
	GetGameGenre(ctx context.Context, gameGenreID values.GameGenreID) (*domain.GameGenre, error)
}

type GameGenreInfo struct {
	domain.GameGenre
	Num int //そのジャンルに含まれるゲームの数
}
