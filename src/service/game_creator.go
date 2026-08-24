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
	// GetGameCreatorJobs
	// ゲームIDに紐づくゲームクリエイターのプリセットジョブ一覧とカスタムジョブ一覧を取得する。
	// 該当するゲームが存在しない場合、ErrInvalidGameIDを返す。
	GetGameCreatorJobs(ctx context.Context, gameID values.GameID) ([]*domain.GameCreatorJob, []*domain.GameCreatorCustomJob, error)
	// DeleteGameCreator
	// game id と creator id に対応するゲームクリエイターを削除する。
	// そのクリエイターとジョブの関連も削除する。
	// 該当するクリエイターが存在しない場合、ErrInvalidGameCreatorIDを返す。
	// クリエイターとゲームが紐づいていない場合は、ErrInvalidGameCreatorGamePairを返す。
	DeleteGameCreator(ctx context.Context, gameID values.GameID, creatorID values.GameCreatorID) error
	// SetGameCreatorJobs
	// game id と creator id に対応するゲームクリエイターのジョブを編集する。
	// 該当の crator の job が jobIDs で置き換えられる。
	//
	// 存在しない creator id の場合は ErrInvalidGameCreatorID を返す。
	// creator id と game id の組み合わせが正しくない場合、ErrInvalidGameCreatorGamePair を返す。
	// 存在しない preset job id の場合、ErrInvalidPresetJobID を返す。
	// その game に紐づいていない custom job id の場合、ErrInvalidGameAndCustomJobPair を返す。
	SetGameCreatorJobs(ctx context.Context, gameID values.GameID, creatorID values.GameCreatorID, jobIDs []values.GameCreatorJobID) error
	// EditGameCreators
	// ゲームクリエイターのジョブを置き換える形で編集する。
	// 該当するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// 存在しないユーザーIDが含まれる場合、 ErrInvalidUserIDを返す。
	// 同じユーザーIDが複数含まれる場合、ErrDuplicateUserIDを返す。
	// 存在しない job ID が含まれる場合、ErrInvalidGameCreatorJobIDを返す。
	// すでに存在するカスタムジョブ名が新しいカスタムジョブとして含まれる場合、ErrDuplicateCustomJobDisplayName を返す。
	// 同一ユーザーに同じjob idが複数含まれる場合は、ErrDuplicateGameCreatorJobIDを返す。
	EditGameCreators(ctx context.Context, session *domain.OIDCSession, gameID values.GameID, inputs []*EditGameCreatorJobInput) error
}

type EditGameCreatorJobInput struct {
	UserID            values.TraPMemberID
	Jobs              []values.GameCreatorJobID
	NewCustomJobNames []values.GameCreatorJobDisplayName
}
