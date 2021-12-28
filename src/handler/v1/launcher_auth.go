package v1

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type LauncherAuth struct {
	launcherAuthService service.LauncherAuth
}

func NewLauncherAuth(launcherAuthService service.LauncherAuth) *LauncherAuth {
	return &LauncherAuth{
		launcherAuthService: launcherAuthService,
	}
}

func (la *LauncherAuth) PostKeyGenerate(c echo.Context, productKeyGen *openapi.ProductKeyGen) ([]*openapi.ProductKey, error) {
	ctx := c.Request().Context()

	keyNum := int(productKeyGen.Num)
	if keyNum < 1 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "num must be greater than 0")
	}

	uuidVersionID, err := uuid.Parse(productKeyGen.Version)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid version")
	}
	versionID := values.NewLauncherVersionIDFromUUID(uuidVersionID)

	launcherUsers, err := la.launcherAuthService.CreateLauncherUser(ctx, versionID, keyNum)
	if errors.Is(err, service.ErrInvalidLauncherVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid version")
	}
	if err != nil {
		log.Printf("error: failed to create launcher user: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to create launcher user")
	}

	productKeys := make([]*openapi.ProductKey, 0, len(launcherUsers))
	for _, launcherUser := range launcherUsers {
		productKeys = append(productKeys, &openapi.ProductKey{
			Key: string(launcherUser.GetProductKey()),
		})
	}

	return productKeys, nil
}

func (la *LauncherAuth) PostLauncherLogin(c echo.Context, productKey *openapi.ProductKey) (*openapi.LauncherAuthToken, error) {
	ctx := c.Request().Context()

	productKeyValue := values.NewLauncherUserProductKeyFromString(productKey.Key)
	err := productKeyValue.Validate()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	launcherSession, err := la.launcherAuthService.LoginLauncher(ctx, productKeyValue)
	if errors.Is(err, service.ErrInvalidLauncherUserProductKey) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid product key")
	}
	if err != nil {
		log.Printf("error: failed to login launcher: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to login launcher")
	}

	return &openapi.LauncherAuthToken{
		AccessToken: string(launcherSession.GetAccessToken()),
		ExpiresIn:   int32(time.Until(launcherSession.GetExpiresAt()).Seconds()),
	}, nil
}

func (la *LauncherAuth) DeleteProductKey(c echo.Context, productKeyID string) error {
	ctx := c.Request().Context()

	uuidLauncherUserID, err := uuid.Parse(productKeyID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid product key id")
	}
	launcherUserID := values.NewLauncherUserIDFromUUID(uuidLauncherUserID)

	err = la.launcherAuthService.RevokeProductKey(ctx, launcherUserID)
	if errors.Is(err, service.ErrInvalidLauncherUser) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid product key")
	}
	if err != nil {
		log.Printf("error: failed to delete launcher user: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete launcher user")
	}

	return nil
}

func (la *LauncherAuth) GetProductKeys(c echo.Context, launcherVersionID string) ([]*openapi.ProductKeyDetail, error) {
	ctx := c.Request().Context()

	uuidLauncherVersionID, err := uuid.Parse(launcherVersionID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid launcher version id")
	}
	launcherVersionIDValue := values.NewLauncherVersionIDFromUUID(uuidLauncherVersionID)

	launcherUsers, err := la.launcherAuthService.GetLauncherUsers(ctx, launcherVersionIDValue)
	if errors.Is(err, service.ErrInvalidLauncherVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid launcher version")
	}
	if err != nil {
		log.Printf("error: failed to get launcher users: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get launcher users")
	}

	productKeys := make([]*openapi.ProductKeyDetail, 0, len(launcherUsers))
	for _, launcherUser := range launcherUsers {
		productKeys = append(productKeys, &openapi.ProductKeyDetail{
			Id:  uuid.UUID(launcherUser.GetID()).String(),
			Key: string(launcherUser.GetProductKey()),
		})
	}

	return productKeys, nil
}

func (la *LauncherAuth) GetLauncherMe(c echo.Context) (*openapi.Version, error) {
	launcherVersion, err := getLauncherVersion(c)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get launcher version")
	}

	var strQuestionnaireURL string
	questionnaireURL, err := launcherVersion.GetQuestionnaireURL()
	if errors.Is(err, domain.ErrNoQuestionnaire) {
		strQuestionnaireURL = ""
	} else if err != nil {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
	} else {
		strQuestionnaireURL = (*url.URL)(questionnaireURL).String()
	}

	return &openapi.Version{
		Id:        uuid.UUID(launcherVersion.GetID()).String(),
		Name:      string(launcherVersion.GetName()),
		AnkeTo:    strQuestionnaireURL,
		CreatedAt: launcherVersion.GetCreatedAt(),
	}, nil
}
