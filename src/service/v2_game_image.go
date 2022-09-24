package service

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

import (
	"context"
	"io"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

type GameImageV2 interface {
	// SaveGameImage
	// ゲーム画像の保存。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	SaveGameImage(ctx context.Context, reader io.Reader, gameID values.GameID) error
	// GetGameImage
	// ゲーム画像一覧の取得。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	GetGameImages(ctx context.Context, gameID values.GameID) ([]*domain.GameImage, error)
	// GetGameImage
	// ゲーム画像の一時的(1分間)に有効なurlを返す。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// ゲーム画像IDに対応するゲーム画像が存在しない、
	// もしくは存在しても紐づくゲームのゲームIDが異なる場合、ErrInvalidGameImageIDを返す。
	GetGameImage(ctx context.Context, gameID values.GameID, imageID values.GameImageID) (values.GameImageTmpURL, error)
	// GetGameImageMeta
	// ゲーム画像のメタデータの取得。
	// ゲームIDに対応するゲームが存在しない場合、ErrInvalidGameIDを返す。
	// ゲーム画像IDに対応するゲーム画像が存在しない場合、ErrInvalidGameImageIDを返す。
	GetGameImageMeta(ctx context.Context, gameID values.GameID, imageID values.GameImageID) (*domain.GameImage, error)
}
