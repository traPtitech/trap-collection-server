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

	//GetGames
	//取得する個数の上限(limit>=0)と開始位置(offset>=0)を指定してゲームを取得する。
	//上限なしはlimit=0。返り値のintは制限をかけないときのゲーム数で、エラーのときは0。また、offsetのみを指定することはできない。
	//limitが負のとき、ErrNegativeLimitを返す。
	//offsetを指定してlimitを設定しなかったとき、ErrOffsetWithoutLimitを返すが、これはserviceで止める。
	//それ以外limitとoffsetがまずかった場合、ErrBadLimitAndOffsetを返す。
	GetGames(ctx context.Context, limit int, offset int) ([]*domain.Game, int, error)

	//GetGamesByUser
	//ユーザーのuuidと取得する個数の上限(limit)と開始位置(offset)を指定して、その人が作成したゲームを取得する。
	//上限なしはlimit=0。返り値のintは制限をかけないときのその人が作ったゲーム数で、エラーのときは0また、offsetのみを指定することはできない。
	//limitが負のとき、ErrNegativeLimitを返す。
	//offsetを指定してlimitを設定しなかったとき、ErrOffsetWithoutLimitを返すが、これはserviceで止める。
	//それ以外limitとoffsetがまずかった場合、ErrBadLimitAndOffsetを返す。
	GetGamesByUser(ctx context.Context, userID values.TraPMemberID, limit int, offset int) ([]*domain.Game, int, error)
}
