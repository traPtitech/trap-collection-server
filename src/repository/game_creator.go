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
	// GetGameCreatorPresetJobs
	// あらかじめ用意されている、プリセットのゲームクリエイターのジョブ一覧を取得する
	GetGameCreatorPresetJobs(ctx context.Context) ([]*domain.GameCreatorJob, error)
	// GetGameCreatorCustomJobsByGameID
	// ゲームIDに紐づくカスタムゲームクリエイターのジョブ一覧を取得する
	GetGameCreatorCustomJobsByGameID(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorCustomJob, error)
	// CreateGameCreatorCustomJobs
	// custom jobを作成する
	CreateGameCreatorCustomJobs(ctx context.Context, customJobs []*domain.GameCreatorCustomJob) error
	// CreateGameCreators
	// ゲームクリエイターを作成する
	CreateGameCreators(ctx context.Context, creators []*domain.GameCreator) error
	// UpsertGameCreatorPresetJobsRelations
	// creator と preset job の relation を更新する
	UpsertGameCreatorPresetJobsRelations(ctx context.Context, jobs map[values.GameCreatorID][]values.GameCreatorJobID) error
	// UpsertGameCreatorCustomJobsRelations
	// creator と custom job の relation を更新する
	UpsertGameCreatorCustomJobsRelations(ctx context.Context, jobs map[values.GameCreatorID][]values.GameCreatorJobID) error
	// GetCreatorsByUserIDs
	// ユーザーIDに紐づくゲームクリエイターを取得する
	GetCreatorsByUserIDs(ctx context.Context, userIDs []values.TraPMemberID) ([]*domain.GameCreator, error)
}
