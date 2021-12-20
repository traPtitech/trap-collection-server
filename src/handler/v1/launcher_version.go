package v1

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
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

func (lv *LauncherVersion) PostVersion(newVersion *openapi.NewVersion) (*openapi.VersionMeta, error) {
	ctx := context.Background()

	name := values.NewLauncherVersionName(newVersion.Name)

	err := name.Validate()
	if errors.Is(err, values.ErrLauncherVersionNameEmpty) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "name is empty")
	}
	if errors.Is(err, values.ErrLauncherVersionNameTooLong) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "name is too long")
	}
	if err != nil {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to validate name")
	}

	var questionnaireURL values.LauncherVersionQuestionnaireURL
	if len(newVersion.AnkeTo) != 0 {
		urlQuestionnaireURL, err := url.Parse(newVersion.AnkeTo)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid questionnaire url")
		}

		questionnaireURL = values.NewLauncherVersionQuestionnaireURL(urlQuestionnaireURL)
	}

	version, err := lv.launcherVersionService.CreateLauncherVersion(ctx, name, questionnaireURL)
	if err != nil {
		log.Printf("error: failed to create launcher version: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to create launcher version")
	}

	var strQuestionnaireURL string
	questionnaireURL, err = version.GetQuestionnaireURL()
	if errors.Is(err, domain.ErrNoQuestionnaire) {
		strQuestionnaireURL = ""
	} else if err != nil {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
	} else {
		strQuestionnaireURL = (*url.URL)(questionnaireURL).String()
	}

	return &openapi.VersionMeta{
		Id:        uuid.UUID(version.GetID()).String(),
		Name:      string(version.GetName()),
		AnkeTo:    strQuestionnaireURL,
		CreatedAt: version.GetCreatedAt(),
	}, nil
}

func (lv *LauncherVersion) GetVersion(strLauncherVersionID string) (*openapi.VersionDetails, error) {
	ctx := context.Background()

	uuidLauncherVersionID, err := uuid.Parse(strLauncherVersionID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid launcher version id")
	}

	launcherVersionID := values.NewLauncherVersionIDFromUUID(uuidLauncherVersionID)

	version, games, err := lv.launcherVersionService.GetLauncherVersion(ctx, launcherVersionID)
	if errors.Is(err, service.ErrNoLauncherVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no such launcher version")
	}
	if err != nil {
		log.Printf("error: failed to get launcher version: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get launcher version")
	}

	var strQuestionnaireURL string
	questionnaireURL, err := version.GetQuestionnaireURL()
	if errors.Is(err, domain.ErrNoQuestionnaire) {
		strQuestionnaireURL = ""
	} else if err != nil {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
	} else {
		strQuestionnaireURL = (*url.URL)(questionnaireURL).String()
	}

	apiGames := make([]openapi.GameMeta, 0, len(games))
	for _, game := range games {
		apiGames = append(apiGames, openapi.GameMeta{
			Id:   uuid.UUID(game.GetID()).String(),
			Name: string(game.GetName()),
		})
	}

	return &openapi.VersionDetails{
		Id:        uuid.UUID(version.GetID()).String(),
		Name:      string(version.GetName()),
		AnkeTo:    strQuestionnaireURL,
		CreatedAt: version.GetCreatedAt(),
		Games:     apiGames,
	}, nil
}

func (lv *LauncherVersion) PostGameToVersion(strLauncherVersionID string, apiGameIDs *openapi.GameIDs) (*openapi.VersionDetails, error) {
	ctx := context.Background()

	uuidLauncherVersionID, err := uuid.Parse(strLauncherVersionID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid launcher version id")
	}

	gameIDs := make([]values.GameID, 0, len(apiGameIDs.GameIDs))
	for _, strGameID := range apiGameIDs.GameIDs {
		gameID, err := uuid.Parse(strGameID)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid game id(%s)", strGameID))
		}

		gameIDs = append(gameIDs, values.NewGameIDFromUUID(gameID))
	}

	launcherVersion, games, err := lv.launcherVersionService.AddGamesToLauncherVersion(
		ctx,
		values.NewLauncherVersionIDFromUUID(uuidLauncherVersionID),
		gameIDs,
	)
	if errors.Is(err, service.ErrNoLauncherVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no such launcher version")
	}
	if errors.Is(err, service.ErrNoGame) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no such game")
	}
	if errors.Is(err, service.ErrDuplicateGame) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "duplicate game")
	}
	if err != nil {
		log.Printf("error: failed to add games to launcher version: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to add games to launcher version")
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

	apiGames := make([]openapi.GameMeta, 0, len(games))
	for _, game := range games {
		apiGames = append(apiGames, openapi.GameMeta{
			Id:   uuid.UUID(game.GetID()).String(),
			Name: string(game.GetName()),
		})
	}

	return &openapi.VersionDetails{
		Id:        uuid.UUID(launcherVersion.GetID()).String(),
		Name:      string(launcherVersion.GetName()),
		AnkeTo:    strQuestionnaireURL,
		CreatedAt: launcherVersion.GetCreatedAt(),
		Games:     apiGames,
	}, nil
}
