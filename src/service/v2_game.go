package service

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

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
	// 存在しないジャンルの場合はジャンルが新規作成される。ジャンルの重複はErrDuplicateGameGenreを返す。
	CreateGame(ctx context.Context, session *domain.OIDCSession, name values.GameName, description values.GameDescription, visibility values.GameVisibility, owners []values.TraPMemberName, maintainers []values.TraPMemberName, gameGenreNames []values.GameGenreName) (*GameInfoV2, error)

	// GetGame
	// ゲームのidを指定してゲーム（id、名前、説明、オーナー、メンテナー）を取得する。
	// idが一致するゲームが存在しなかった場合、ErrNoGameを返す。
	// sessionがnilの場合、返り値のOwnersとMaintainersは空配列。
	GetGame(ctx context.Context, session *domain.OIDCSession, gameID values.GameID) (*GameInfoV2, error)

	// GetGames
	// ゲームにいろいろ制限をかけて取得。limitは取得上限、offsetは取得開始位置。limitが0のときは全て取得。
	// visibilitiesは取得するゲームの公開範囲を指定。空配列の場合は全ての公開範囲のゲームを取得。
	// gameGenreIDsは取得するゲームのジャンルを指定。空配列の場合は全てのジャンルのゲームを取得。
	// gameNameはゲーム名の部分一致検索。空文字の場合は全てのゲームを取得。
	// 返り値のintは制限をかけない場合のゲーム数
	// sortTypeがおかしいときErrInvalidGamesSortType
	GetGames(
		ctx context.Context, limit int, offset int, sort GamesSortType,
		visibilities []values.GameVisibility, gameGenreIDs []values.GameGenreID, gameName string) (int, []*domain.GameWithGenres, error)

	// GetMyGames
	// ログイン中のユーザーが作ったゲームを制限をかけて取得。limitは取得上限、offsetは取得開始位置。limitが0のときは全て取得。
	// visibilitiesは取得するゲームの公開範囲を指定。空配列の場合は全ての公開範囲のゲームを取得。
	// gameGenreIDsは取得するゲームのジャンルを指定。空配列の場合は全てのジャンルのゲームを取得。
	// gameNameはゲーム名の部分一致検索。空文字の場合は全てのゲームを取得。
	// 返り値のintは制限をかけない場合のゲーム数
	// sortTypeがおかしいときErrInvalidGamesSortType
	GetMyGames(
		ctx context.Context, session *domain.OIDCSession, limit int, offset int, sort GamesSortType,
		visibilities []values.GameVisibility, gameGenreIDs []values.GameGenreID, gameName string) (int, []*domain.GameWithGenres, error)

	// UpdateGame
	// ゲームのidを指定して情報（名前、説明）を修正する。
	// idが一致するゲームが存在しなかった場合、ErrNoGameを返す。
	// visibilityが変わらないときはnilを指定する。
	UpdateGame(ctx context.Context, gameID values.GameID, name values.GameName, description values.GameDescription, visibility *values.GameVisibility) (*domain.Game, error)

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

type GamesSortType int

const (
	GamesSortTypeCreatedAt GamesSortType = iota
	GamesSortTypeLatestVersion
)
