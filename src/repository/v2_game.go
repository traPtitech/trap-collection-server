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
	//ゲームの情報（名前や説明など）を書き換える
	UpdateGame(ctx context.Context, game *domain.Game) error

	//RemoveGame
	//指定されたidのゲームを削除する
	RemoveGame(ctx context.Context, gameID values.GameID) error

	//GetGame
	//指定されたidのゲームを取得する
	GetGame(ctx context.Context, gameID values.GameID, lockType LockType) (*domain.Game, error)

	//GetGames
	//取得する個数の上限(limit)と開始位置(offset)を指定してゲームを取得する。
	//上限なしはlimit=-1。返り値のintは制限をかけないときのゲーム数
	GetGames(ctx context.Context, limit int, offset int) ([]*domain.Game, int, error)

	//GetGamesByUser
	//ユーザーのuuidと取得する個数の上限(limit)と開始位置(offset)を指定して、その人が作成したゲームを取得する。
	//上限なしはlimit=-1。返り値のintは制限をかけないときのその人が作ったゲーム数
	GetGamesByUser(ctx context.Context, userID values.TraPMemberID, limit int, offset int) ([]*domain.Game, int, error)
}