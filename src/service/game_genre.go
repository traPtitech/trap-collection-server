package service

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre interface {
	// 全てのゲームジャンルとそのジャンルのゲーム数を取得する。
	// ユーザーがログインしていない場合、privateなゲームの数は含まれない。
	GetGameGenres(ctx context.Context, isLoginUser bool) ([]*GameGenreInfo, error)
	// ゲームジャンルを削除する。
	// ゲームジャンルが存在しない場合は、ErrNoGameGenreを返す。
	DeleteGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error
	// ゲームが持つジャンルを修正する。
	// ゲームが存在しない場合は、ErrNoGameを返す。
	// ジャンル名に重複がある場合は、ErrDuplicateGameGenreを返す。
	// 存在しないジャンル名が含まれている場合は、ジャンルを作成する。
	UpdateGameGenres(ctx context.Context, gameID values.GameID, gameGenreNames []values.GameGenreName) error
	// ゲームジャンルの情報を編集する。
	// ゲームジャンルが存在しない場合は、ErrNoGameGenreを返す。
	// ジャンル名に重複がある場合は、ErrDuplicateGameGenreNameを返す。
	// ジャンル名が変更されなかった場合は、ErrNoGameGenreUpdatedを返す。
	UpdateGameGenre(ctx context.Context, gameGenreID values.GameGenreID, gameGenreName values.GameGenreName) (*GameGenreInfo, error)
}

type GameGenreInfo struct {
	domain.GameGenre
	Num int // そのジャンルを持つゲームの数
}
