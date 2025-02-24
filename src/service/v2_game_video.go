package service

//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameVideoV2 interface {
	// SaveGameVideo
	// ゲーム動画の保存。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	SaveGameVideo(ctx context.Context, reader io.Reader, gameID values.GameID) (*domain.GameVideo, error)
	// GetGameVideos
	// ゲーム動画一覧の取得。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	GetGameVideos(ctx context.Context, gameID values.GameID) ([]*domain.GameVideo, error)
	// GetGameVideo
	// ゲーム動画の一時的(1分間)に有効なurlを返す。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// ゲーム動画IDに対応するゲーム動画が存在しない場合、
	// もしくは存在しても紐づくゲームのゲームIDが異なる場合、ErrInvalidGameVideoIDを返す。
	GetGameVideo(ctx context.Context, gameID values.GameID, videoID values.GameVideoID) (values.GameVideoTmpURL, error)
	// GetGameVideoMeta
	// ゲーム動画のメタデータの取得。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// ゲーム動画IDに対応するゲーム動画が存在しない場合、ErrInvalidGameVideoIDを返す。
	GetGameVideoMeta(ctx context.Context, gameID values.GameID, videoID values.GameVideoID) (*domain.GameVideo, error)
}
