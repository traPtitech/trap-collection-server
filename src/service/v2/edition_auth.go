package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/repository"
	"github.com/traPtitech/trap-collection-server/src/service"
)

var _ service.EditionAuth = (*EditionAuth)(nil)

const expiresIn = 86400

type EditionAuth struct {
	db                    repository.DB
	editionRepository     repository.Edition
	productKeyRepository  repository.ProductKey
	accessTokenRepository repository.AccessToken
}

func NewEditionAuth(
	db repository.DB,
	editionRepository repository.Edition,
	productKeyRepository repository.ProductKey,
	accessTokenRepository repository.AccessToken,
) *EditionAuth {
	return &EditionAuth{
		db:                    db,
		editionRepository:     editionRepository,
		productKeyRepository:  productKeyRepository,
		accessTokenRepository: accessTokenRepository,
	}
}

func (editionAuth *EditionAuth) GenerateProductKey(ctx context.Context, editionID values.LauncherVersionID, num uint) ([]*domain.LauncherUser, error) {
	if num == 0 {
		return nil, service.ErrInvalidKeyNum
	}

	_, err := editionAuth.editionRepository.GetEdition(ctx, editionID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidLauncherVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	productKeys := make([]*domain.LauncherUser, 0, num)
	for i := uint(0); i < num; i++ {
		productKey, err := values.NewLauncherUserProductKey()
		if err != nil {
			return nil, fmt.Errorf("failed to create product key: %w", err)
		}

		productKeys = append(productKeys, domain.NewLauncherUser(
			values.NewLauncherUserID(),
			productKey,
		))
	}

	err = editionAuth.productKeyRepository.SaveProductKeys(ctx, editionID, productKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher users: %w", err)
	}

	return productKeys, nil
}

func (editionAuth *EditionAuth) GetProductKeys(ctx context.Context, editionID values.LauncherVersionID) ([]*domain.LauncherUser, error) {
	_, err := editionAuth.editionRepository.GetEdition(ctx, editionID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidLauncherVersion
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher version: %w", err)
	}

	productKeys, err := editionAuth.productKeyRepository.GetProductKeys(ctx, editionID, repository.LockTypeNone)
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher users: %w", err)
	}

	return productKeys, nil
}

func (editionAuth *EditionAuth) ActivateProductKey(ctx context.Context, productKeyID values.LauncherUserID) (*domain.LauncherUser, error) {
	productKey, err := editionAuth.productKeyRepository.GetProductKey(ctx, productKeyID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidProductKey
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher user: %w", err)
	}

	if productKey.GetStatus() == values.LauncherUserStatusActive {
		return nil, service.ErrKeyAlreadyActivated
	}

	productKey.SetStatus(values.LauncherUserStatusActive)

	err = editionAuth.productKeyRepository.UpdateProductKey(ctx, productKey)
	if err != nil {
		return nil, fmt.Errorf("failed to delete launcher user: %w", err)
	}

	return productKey, nil
}

func (editionAuth *EditionAuth) RevokeProductKey(ctx context.Context, productKeyID values.LauncherUserID) (*domain.LauncherUser, error) {
	productKey, err := editionAuth.productKeyRepository.GetProductKey(ctx, productKeyID, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidProductKey
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher user: %w", err)
	}

	if productKey.GetStatus() == values.LauncherUserStatusInactive {
		return nil, service.ErrKeyAlreadyRevoked
	}

	productKey.SetStatus(values.LauncherUserStatusInactive)

	err = editionAuth.productKeyRepository.UpdateProductKey(ctx, productKey)
	if err != nil {
		return nil, fmt.Errorf("failed to delete launcher user: %w", err)
	}

	return productKey, nil
}

func (editionAuth *EditionAuth) AuthorizeEdition(ctx context.Context, key values.LauncherUserProductKey) (*domain.LauncherSession, error) {
	productKey, err := editionAuth.productKeyRepository.GetProductKeyByKey(ctx, key)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, service.ErrInvalidProductKey
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get launcher user: %w", err)
	}

	if productKey.GetStatus() != values.LauncherUserStatusActive {
		return nil, service.ErrInvalidProductKey
	}

	token, err := values.NewLauncherSessionAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	accessToken := domain.NewLauncherSession(
		values.NewLauncherSessionID(),
		token,
		getExpiresAt(),
	)

	err = editionAuth.accessTokenRepository.SaveAccessToken(ctx, productKey.GetID(), accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create launcher session: %w", err)
	}

	return accessToken, nil
}

func getExpiresAt() time.Time {
	return time.Now().Add(expiresIn * time.Second)
}

func (editionAuth *EditionAuth) EditionAuth(ctx context.Context, token values.LauncherSessionAccessToken) (*domain.LauncherUser, *domain.LauncherVersion, error) {
	accessTokenInfo, err := editionAuth.accessTokenRepository.GetAccessTokenInfo(ctx, token, repository.LockTypeNone)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, nil, service.ErrInvalidAccessToken
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get launcher version and user and session: %w", err)
	}

	if accessTokenInfo.ProductKey.GetStatus() == values.LauncherUserStatusInactive {
		return nil, nil, service.ErrInvalidAccessToken
	}

	if accessTokenInfo.AccessToken.IsExpired() {
		return nil, nil, service.ErrExpiredAccessToken
	}

	return accessTokenInfo.ProductKey, accessTokenInfo.Edition, nil
}
