package repository

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock -typed

type GameCreator interface {
	// GetGameCreatorsByGameID
	// ゲームIDに紐づくゲームクリエイターとそのジョブ一覧を取得する
	GetGameCreatorsByGameID(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorWithJobs, error)
}
