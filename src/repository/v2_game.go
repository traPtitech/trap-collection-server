package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameV2 interface {
	//SaveGame
	//ゲームを保存する
	SaveGame(ctx context.Context, game *domain.Game) error

	//UpdateGame
	//ゲームの情報（名前や説明など）を書き換える。
	//ゲームが存在しなかったとき、ErrNoRecordUpdatedを返す。
	UpdateGame(ctx context.Context, game *domain.Game) error

	//RemoveGame
	//指定されたidのゲームを削除する
	//ゲームが存在しなかったとき、ErrNoRecordDeletedを返す。
	RemoveGame(ctx context.Context, gameID values.GameID) error

	//GetGame
	//指定されたidのゲームを取得する。
	//ゲームが見つからなかったとき、ErrRecordNotFoundを返す。
	GetGame(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.Game, error)

	//GetGamesByIDs
	//指定されたidのゲームを取得する。
	GetGamesByIDs(ctx context.Context, gameIDs []values.GameID, lockType LockType) ([]*domain.Game, error)

	// GetGames
	// 取得する個数の上限(limit>=0)と開始位置(offset>=0)を指定してゲームを取得する。
	// 上限なしはlimit=0。返り値のintは制限をかけないときのゲーム数で、エラーのときは0。また、offsetのみを指定することはできない。
	// limitが負のとき、ErrNegativeLimitを返す。
	// visibilitiesが無いときは、全てのゲームを取得する。
	// userIDが指定されているときは、そのユーザーが作成したゲームを取得する。
	// gameGenresが指定されているときは、そのジャンルがすべて含まれるゲームを取得する。
	// nameが指定されているときは、その名前を含むゲームを取得する。
	GetGames(
		// 必須
		ctx context.Context,
		limit int,
		offset int,
		sort GamesSortType,
		// nil、空文字列でもよい
		visibilities []values.GameVisibility,
		userID *values.TraPMemberID,
		gameGenres []values.GameGenreID,
		name string,
	) ([]*domain.GameWithGenres, int, error)
}

type GamesSortType int

const (
	GamesSortTypeCreatedAt GamesSortType = iota
	GamesSortTypeLatestVersion
)
