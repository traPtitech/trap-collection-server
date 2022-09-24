package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//GameV2
type GameV2 interface {
	//CreateGame
	//ゲームの追加。
	//v1ではセッションを使っていたがv2には無い
	CreateGame(ctx context.Context, name values.GameName, description values.GameDescription, owners []values.TraPMemberName, maintainers []values.TraPMemberName) (*GameInfoV2, error)

	//GetGame
	//ゲームのidを指定してゲーム（id、名前、説明、オーナー、メンテナー）を取得する。
	GetGame(ctx context.Context, gameID values.GameID) (*GameInfoV2, error)

	//GetGames
	//ゲームを全部取得。
	GetGames(ctx context.Context) (*GameInfoV2, error)

	//UpdateGame
	//ゲームのidを指定して情報（名前、説明）を修正する。
	UpdateGame(ctx context.Context, gameID values.GameID, name values.GameName, description values.GameDescription) (*domain.Game, error)

	//DeleteGame
	//ゲームのidを指定してゲームを削除する。
	DeleteGame(ctx context.Context, gameID values.GameID) error
}

//GameInfoV2(struct)
//V2になってゲームバージョンを返さなくなり、オーナーとメンテナーを返すようになったので追加。
type GameInfoV2 struct {
	*domain.Game
	owners      []values.TraPMemberName
	mainrainers []values.TraPMemberName
}
