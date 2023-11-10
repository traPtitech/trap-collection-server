package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

// GameV2
type GameV2 interface {
	// CreateGame
	// ゲームの追加。
	// owners内に重複がある場合、ErrOverlapInOwnersを返す。
	// maintainers内に重複がある場合、ErrOverlapInMaintainersを返す。
	// ownersとmaintainersに重複がある場合、また、ログイン中のユーザーがmaintainersに含まれる場合、ErrOverlapBetweenOwnersAndMaintainersを返す。
	CreateGame(ctx context.Context, session *domain.OIDCSession, name values.GameName, description values.GameDescription, visibility values.GameVisibility, owners []values.TraPMemberName, maintainers []values.TraPMemberName) (*GameInfoV2, error)

	// GetGame
	// ゲームのidを指定してゲーム（id、名前、説明、オーナー、メンテナー）を取得する。
	// idが一致するゲームが存在しなかった場合、ErrNoGameを返す。
	GetGame(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) (*GameInfoV2, error)

	// GetGames
	// ゲームにいろいろ制限をかけて取得。limitは取得上限、offsetは取得開始位置。上限なしで取得する場合limit=0。
	// 返り値のintは制限をかけない場合のゲーム数
	GetGames(ctx context.Context, limit int, offset int) (int, []*domain.Game, error)

	// GetMyGames
	// ログイン中のユーザーが作ったゲームを制限をかけて取得。limitは取得上限、offsetは取得開始位置。上限なしで取得する場合limit=0。
	// 返り値のintは制限をかけない場合のゲーム数
	GetMyGames(ctx context.Context, session *domain.OIDCSession, limit int, offset int) (int, []*domain.Game, error)

	// UpdateGame
	// ゲームのidを指定して情報（名前、説明）を修正する。
	// idが一致するゲームが存在しなかった場合、ErrNoGameを返す。
	UpdateGame(ctx context.Context, gameID values.GameID, name values.GameName, description values.GameDescription) (*domain.Game, error)

	// DeleteGame
	// ゲームのidを指定してゲームを削除する。
	// idが一致するゲームが存在しなかった場合、ErrNoGameを返す。
	DeleteGame(ctx context.Context, gameID values.GameID) error
}

// GameInfoV2(struct)
// V2になってゲームバージョンを返さなくなり、オーナーとメンテナーを返すようになったので追加。
type GameInfoV2 struct {
	Game        *domain.Game
	Owners      []*UserInfo
	Maintainers []*UserInfo
	Genres      []*domain.GameGenre
}
