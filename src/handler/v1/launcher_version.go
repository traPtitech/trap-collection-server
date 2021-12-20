package v1

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type LauncherVersion struct {
	launcherVersionService service.LauncherVersion
}

func NewLauncherVersion(launcherVersionService service.LauncherVersion) *LauncherVersion {
	return &LauncherVersion{
		launcherVersionService: launcherVersionService,
	}
}

func (lv *LauncherVersion) GetVersions() ([]*openapi.Version, error) {
	ctx := context.Background()

	versions, err := lv.launcherVersionService.GetLauncherVersions(ctx)
	if err != nil {
		log.Printf("error: failed to get launcher versions: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get launcher versions")
	}

	apiVersions := make([]*openapi.Version, 0, len(versions))
	for _, version := range versions {
		var questionnaire string
		questionnaireURL, err := version.GetQuestionnaireURL()
		if errors.Is(err, domain.ErrNoQuestionnaire) {
			questionnaire = ""
		} else if err != nil {
			log.Printf("error: failed to get questionnaire url: %v\n", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
		} else {
			questionnaire = (*url.URL)(questionnaireURL).String()
		}

		apiVersions = append(apiVersions, &openapi.Version{
			Id:        uuid.UUID(version.GetID()).String(),
			Name:      string(version.GetName()),
			AnkeTo:    questionnaire,
			CreatedAt: version.GetCreatedAt(),
		})
	}

	return apiVersions, nil
}
