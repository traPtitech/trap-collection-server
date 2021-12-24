package v1

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
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

func (lv *LauncherVersion) GetCheckList(operatingSystem string, c echo.Context) ([]*openapi.CheckItem, error) {
	launcherVersion, err := getLauncherVersion(c)
	if err != nil {
		log.Printf("error: failed to get launcher version: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get launcher version")
	}

	var launcherEnvOS values.LauncherEnvironmentOS
	switch operatingSystem {
	case "win32":
		launcherEnvOS = values.LauncherEnvironmentOSWindows
	case "darwin":
		launcherEnvOS = values.LauncherEnvironmentOSMac
	default:
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid operating system")
	}

	checkList, err := lv.launcherVersionService.GetLauncherVersionCheckList(c.Request().Context(), launcherVersion.GetID(), values.NewLauncherEnvironment(launcherEnvOS))
	if errors.Is(err, service.ErrNoLauncherVersion) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no such launcher version")
	}
	if err != nil {
		log.Printf("error: failed to get launcher version check list: %v\n", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get launcher version check list")
	}

	apiCheckList := make([]*openapi.CheckItem, 0, len(checkList))
	for _, checkItem := range checkList {
		var hash string
		var bodyUpdatedAt time.Time
		var gameType string
		var entryPoint string
		if checkItem.LatestFile != nil {
			switch checkItem.LatestFile.GetFileType() {
			case values.GameFileTypeJar:
				gameType = "jar"
			case values.GameFileTypeWindows:
				gameType = "windows"
			case values.GameFileTypeMac:
				gameType = "mac"
			default:
				log.Printf("error: unknown game file type(gameID: %s): %d\n", uuid.UUID(checkItem.GetID()).String(), checkItem.LatestFile.GetFileType())
				continue
			}

			hash = hex.EncodeToString([]byte(checkItem.LatestFile.GetHash()))
			bodyUpdatedAt = checkItem.LatestFile.GetCreatedAt()
			entryPoint = string(checkItem.LatestFile.GetEntryPoint())
		} else if checkItem.LatestURL != nil {
			gameType = "url"

			bodyUpdatedAt = checkItem.LatestURL.GetCreatedAt()
		} else {
			log.Printf("error: no game file or url specified(gameID: %s)\n", uuid.UUID(checkItem.GetID()).String())
			continue
		}

		if checkItem.LatestImage == nil {
			log.Printf("error: no image or video specified(gameID: %s)\n", uuid.UUID(checkItem.GetID()).String())
			continue
		}

		var movieUpdatedAt time.Time
		if checkItem.LatestVideo != nil {
			movieUpdatedAt = checkItem.LatestVideo.GetCreatedAt()
		}

		apiCheckList = append(apiCheckList, &openapi.CheckItem{
			Id:             uuid.UUID(checkItem.Game.GetID()).String(),
			Md5:            hash,
			Type:           gameType,
			EntryPoint:     entryPoint,
			BodyUpdatedAt:  bodyUpdatedAt,
			ImgUpdatedAt:   checkItem.LatestImage.GetCreatedAt(),
			MovieUpdatedAt: movieUpdatedAt,
		})
	}

	return apiCheckList, nil
}
