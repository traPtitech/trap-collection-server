package service

import (
	"context"

	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
)

//go:generate go tool mockgen -source=$GOFILE -destination=mock/${GOFILE} -package=mock

type EditionAuth interface {
	// GenerateProductKey
	// 指定したエディションに対して、指定した数のプロダクトキーを生成します。
	// エディションが存在しない場合、ErrInvalidEditionIDを返します。
	// numが0の場合、ErrInvalidKeyNumを返します。
	GenerateProductKey(ctx context.Context, editionID values.LauncherVersionID, num uint) ([]*domain.LauncherUser, error)
	// GetProductKeys
	// 指定したエディションのプロダクトキーを取得します。
	// エディションが存在しない場合、ErrInvalidEditionIDを返します。
	GetProductKeys(ctx context.Context, editionID values.LauncherVersionID, params GetProductKeysParams) ([]*domain.LauncherUser, error)
	// ActivateProductKey
	// 指定したプロダクトキーを有効化します。
	// 存在しないプロダクトキーの場合、ErrInvalidProductKeyを返します。
	// 既に有効なプロダクトキーの場合、ErrKeyAlreadyActivatedを返します。
	ActivateProductKey(ctx context.Context, productKey values.LauncherUserID) (*domain.LauncherUser, error)
	// RevokeProductKey
	// 指定したプロダクトキーを無効化します。
	// 存在しないプロダクトキーの場合、ErrInvalidProductKeyを返します。
	// 既に無効なプロダクトキーの場合、ErrKeyAlreadyRevokedを返します。
	RevokeProductKey(ctx context.Context, productKey values.LauncherUserID) (*domain.LauncherUser, error)
	// AuthorizeEdition
	// プロダクトキーから、エディション情報へのアクセストークンを発行します。
	// 存在しないプロダクトキーの場合、ErrInvalidProductKeyを返します。
	AuthorizeEdition(ctx context.Context, productKey values.LauncherUserProductKey) (*domain.LauncherSession, error)
	// EditionAuth
	// エディション情報へのアクセストークンを検証します。
	// アクセストークンが存在しない、もしくは無効な場合、ErrInvalidAccessTokenを返します。
	// アクセストークンが期限切れの場合、ErrExpiredAccessTokenを返します。
	EditionAuth(ctx context.Context, accessToken values.LauncherSessionAccessToken) (*domain.LauncherUser, *domain.LauncherVersion, error)
	// EditionGameAuth
	// エディション情報へのアクセストークンを検証し、
	// ゲーム情報にアクセスできるかどうかチェックします。
	// アクセストークンが存在しない、もしくは無効な場合、ErrInvalidAccessTokenを返します。
	// アクセストークンが期限切れの場合、ErrExpiredAccessTokenを返します。
	// ゲーム情報にアクセスできない、もしくはゲームが存在しない場合、ErrForbiddenを返します。
	EditionGameAuth(ctx context.Context, accessToken values.LauncherSessionAccessToken, gameID values.GameID) (*domain.LauncherUser, *domain.LauncherVersion, error)
	// EditionImageAuth
	// エディション情報へのアクセストークンを検証し、
	// イメージ情報にアクセスできるかどうかチェックします。
	// アクセストークンが存在しない、もしくは無効な場合、ErrInvalidAccessTokenを返します。
	// アクセストークンが期限切れの場合、ErrExpiredAccessTokenを返します。
	// イメージ情報にアクセスできない、もしくはイメージが存在しない場合、ErrForbiddenを返します。
	EditionImageAuth(ctx context.Context, accessToken values.LauncherSessionAccessToken, imageID values.GameImageID) (*domain.LauncherUser, *domain.LauncherVersion, error)
	// EditionVideoAuth
	// エディション情報へのアクセストークンを検証し、
	// 動画情報にアクセスできるかどうかチェックします。
	// アクセストークンが存在しない、もしくは無効な場合、ErrInvalidAccessTokenを返します。
	// アクセストークンが期限切れの場合、ErrExpiredAccessTokenを返します。
	// 動画情報にアクセスできない、もしくは動画が存在しない場合、ErrForbiddenを返します。
	EditionVideoAuth(ctx context.Context, accessToken values.LauncherSessionAccessToken, videoID values.GameVideoID) (*domain.LauncherUser, *domain.LauncherVersion, error)
	// EditionFileAuth
	// エディション情報へのアクセストークンを検証し、
	// ファイル情報にアクセスできるかどうかチェックします。
	// アクセストークンが存在しない、もしくは無効な場合、ErrInvalidAccessTokenを返します。
	// アクセストークンが期限切れの場合、ErrExpiredAccessTokenを返します。
	// ファイル情報にアクセスできない、もしくはファイルが存在しない場合、ErrForbiddenを返します。
	EditionFileAuth(ctx context.Context, accessToken values.LauncherSessionAccessToken, fileID values.GameFileID) (*domain.LauncherUser, *domain.LauncherVersion, error)
}

type GetProductKeysParams struct {
	// Status
	// プロダクトキーのステータス。
	// 指定しない場合は、全てのステータスのプロダクトキーを取得します。
	Status types.Option[values.LauncherUserStatus]
}
