package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre interface {
	// ゲームジャンルを削除する。
	// ゲームジャンルが存在しない場合は、ErrNoGameGenreを返す。
	DeleteGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error
}
