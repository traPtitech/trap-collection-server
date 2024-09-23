package v2

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/pkg/types"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type EditionAuth struct {
	context            *Context
	editionAuthService service.EditionAuth
}

func NewEditionAuth(context *Context, editionAuth service.EditionAuth) *EditionAuth {
	return &EditionAuth{
		context:            context,
		editionAuthService: editionAuth,
	}
}

// プロダクトキーの一覧の取得
// (GET /editions/{editionID}/keys)
func (editionAuth *EditionAuth) GetProductKeys(c echo.Context, editionID openapi.EditionIDInPath, params openapi.GetProductKeysParams) error {
	var status types.Option[values.LauncherUserStatus]
	if params.Status != nil {
		switch *params.Status {
		case openapi.Active:
			status = types.NewOption(values.LauncherUserStatusActive)
		case openapi.Revoked:
			status = types.NewOption(values.LauncherUserStatusInactive)
		default:
			return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
		}
	}

	productKeys, err := editionAuth.editionAuthService.GetProductKeys(
		c.Request().Context(),
		values.NewLauncherVersionIDFromUUID(editionID),
		service.GetProductKeysParams{
			Status: status,
		},
	)
	if errors.Is(err, service.ErrInvalidEditionID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid editionID")
	}
	if err != nil {
		log.Printf("failed to get product keys: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get product keys")
	}

	res := make([]openapi.ProductKey, 0, len(productKeys))
	for _, productKey := range productKeys {
		var status openapi.ProductKeyStatus
		switch productKey.GetStatus() {
		case values.LauncherUserStatusActive:
			status = openapi.Active
		case values.LauncherUserStatusInactive:
			status = openapi.Revoked
		default:
			log.Printf("error: invalid product key status: %v\n", productKey.GetStatus())
			continue
		}

		res = append(res, openapi.ProductKey{
			Id:        uuid.UUID(productKey.GetID()),
			Key:       string(productKey.GetProductKey()),
			Status:    status,
			CreatedAt: productKey.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, res)
}

// プロダクトキーの生成
// (POST /editions/{editionID}/keys)
func (editionAuth *EditionAuth) PostProductKey(c echo.Context, editionID openapi.EditionIDInPath, params openapi.PostProductKeyParams) error {
	productKey, err := editionAuth.editionAuthService.GenerateProductKey(
		c.Request().Context(),
		values.NewLauncherVersionIDFromUUID(editionID),
		uint(params.Num),
	)
	if errors.Is(err, service.ErrInvalidEditionID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid editionID")
	}
	if errors.Is(err, service.ErrInvalidKeyNum) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid key num")
	}
	if err != nil {
		log.Printf("error: failed to create product key: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create product key")
	}

	res := make([]openapi.ProductKey, 0, len(productKey))
	for _, key := range productKey {
		res = append(res, openapi.ProductKey{
			Id:        uuid.UUID(key.GetID()),
			Key:       string(key.GetProductKey()),
			Status:    openapi.Active,
			CreatedAt: key.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusCreated, res)
}

// プロダクトキーの再有効化
// (POST /editions/{editionID}/keys/{productKeyID}/activate)
func (editionAuth *EditionAuth) PostActivateProductKey(c echo.Context, _ openapi.EditionIDInPath, productKeyID openapi.ProductKeyIDInPath) error {
	productKey, err := editionAuth.editionAuthService.ActivateProductKey(
		c.Request().Context(),
		values.NewLauncherUserIDFromUUID(productKeyID),
	)
	if errors.Is(err, service.ErrInvalidProductKey) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid productKeyID")
	}
	if errors.Is(err, service.ErrKeyAlreadyActivated) {
		return echo.NewHTTPError(http.StatusNotFound, "key already activated")
	}
	if err != nil {
		log.Printf("error: failed to activate product key: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to activate product key")
	}

	return c.JSON(http.StatusOK, openapi.ProductKey{
		Id:        uuid.UUID(productKey.GetID()),
		Key:       string(productKey.GetProductKey()),
		Status:    openapi.Active,
		CreatedAt: productKey.GetCreatedAt(),
	})
}

// プロダクトキーの失効
// (POST /editions/{editionID}/keys/{productKeyID}/revoke)
func (editionAuth *EditionAuth) PostRevokeProductKey(c echo.Context, _ openapi.EditionIDInPath, productKeyID openapi.ProductKeyIDInPath) error {
	productKey, err := editionAuth.editionAuthService.RevokeProductKey(
		c.Request().Context(),
		values.NewLauncherUserIDFromUUID(productKeyID),
	)
	if errors.Is(err, service.ErrInvalidProductKey) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid productKeyID")
	}
	if errors.Is(err, service.ErrKeyAlreadyRevoked) {
		return echo.NewHTTPError(http.StatusNotFound, "key already revoked")
	}
	if err != nil {
		log.Printf("error: failed to revoke product key: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to revoke product key")
	}

	return c.JSON(http.StatusOK, openapi.ProductKey{
		Id:        uuid.UUID(productKey.GetID()),
		Key:       string(productKey.GetProductKey()),
		Status:    openapi.Revoked,
		CreatedAt: productKey.GetCreatedAt(),
	})
}

// ランチャーの認可リクエスト
// (POST /editions/authorize)
func (editionAuth *EditionAuth) PostEditionAuthorize(c echo.Context) error {
	var params openapi.EditionAuthorizeRequest
	err := c.Bind(&params)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	key := values.NewLauncherUserProductKeyFromString(params.Key)
	if err := key.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid product key: %v", err))
	}

	accessToken, err := editionAuth.editionAuthService.AuthorizeEdition(
		c.Request().Context(),
		key,
	)
	if errors.Is(err, service.ErrInvalidProductKey) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid product key")
	}
	if err != nil {
		log.Printf("error: failed to authorize launcher: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to authorize launcher")
	}

	return c.JSON(http.StatusOK, openapi.EditionAccessToken{
		AccessToken: string(accessToken.GetAccessToken()),
		ExpiresAt:   accessToken.GetExpiresAt(),
	})
}

// エディション情報の取得
// (GET /editions/info)
func (editionAuth *EditionAuth) GetEditionInfo(c echo.Context) error {
	edition, err := editionAuth.context.GetEdition(c)
	if err != nil {
		log.Printf("error: failed to get edition: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get edition")
	}

	questionnaireURL, err := edition.GetQuestionnaireURL()
	if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
	}

	var strQuestionnaireURL *string
	if !errors.Is(err, domain.ErrNoQuestionnaire) {
		v := (*url.URL)(questionnaireURL).String()
		strQuestionnaireURL = &v
	}

	return c.JSON(http.StatusOK, openapi.Edition{
		Id:            uuid.UUID(edition.GetID()),
		Name:          string(edition.GetName()),
		Questionnaire: strQuestionnaireURL,
		CreatedAt:     edition.GetCreatedAt(),
	})
}
