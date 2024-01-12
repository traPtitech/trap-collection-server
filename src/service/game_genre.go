package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre interface {
	// 全てのゲームジャンルとそのジャンルのゲーム数を取得する。
	GetGameGenres(ctx context.Context) ([]*GameGenreInfo, error)
	// ゲームジャンルを削除する。
	// ゲームジャンルが存在しない場合は、ErrNoGameGenreを返す。
	DeleteGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error
}

type GameGenreInfo struct {
	domain.GameGenre
	Num int // そのジャンルを持つゲームの数
}
