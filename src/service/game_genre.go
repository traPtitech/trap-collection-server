package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre interface {
	// ゲームジャンルを削除する。
	// ゲームジャンルが存在しない場合は、ErrNoGameGenreを返す。
	DeleteGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error
}
