package repository

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameGenre interface {
	GetGenresByGameID(ctx context.Context, gameID values.GameID) ([]*domain.GameGenre, error)
	// RemoveGameGenre
	// ゲームジャンルを削除する。
	// IDに該当するゲームジャンルが存在しない場合は、ErrNoRecordDeletedを返す。
	RemoveGameGenre(ctx context.Context, gameGenreID values.GameGenreID) error
}
