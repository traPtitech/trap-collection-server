package v1

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type LauncherAuth struct {
	launcherAuthService service.LauncherAuth
	openapi.LauncherAuthApi
}

func NewLauncherAuth(launcherAuthService service.LauncherAuth) *LauncherAuth {
	return &LauncherAuth{
		launcherAuthService: launcherAuthService,
	}
}

func (la *LauncherAuth) PostKeyGenerate(productKeyGen *openapi.ProductKeyGen) ([]*openapi.ProductKey, error) {
	ctx := context.Background()

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
