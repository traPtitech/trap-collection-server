package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock -typed

type GameCreator interface {
	// GetGameCreators
	// ゲームIDに紐づくゲームクリエイターとそのジョブ一覧を取得する。
	// 該当するゲームが存在しない場合、ErrInvalidGameIDを返す。
	GetGameCreators(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorWithJobs, error)
}
