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

type Edition struct {
	editionService service.Edition
}

func NewEdition(editionService service.Edition) *Edition {
	return &Edition{
		editionService: editionService,
	}
}

// エディション一覧の取得
// (GET /editions)
func (edition *Edition) GetEditions(c echo.Context) error {
	editions, err := edition.editionService.GetEditions(c.Request().Context())
	if err != nil {
		log.Printf("error: failed to get editions: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get editions")
	}

	res := make([]openapi.Edition, 0, len(editions))
	for _, edition := range editions {
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

		res = append(res, openapi.Edition{
			Id:            uuid.UUID(edition.GetID()),
			Name:          string(edition.GetName()),
			Questionnaire: strQuestionnaireURL,
			CreatedAt:     edition.GetCreatedAt(),
		})
	}

	return c.JSON(http.StatusOK, res)
}

// エディションの作成
// (POST /editions)
func (edition *Edition) PostEdition(c echo.Context) error {
	var req openapi.NewEdition
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	name := values.NewLauncherVersionName(req.Name)
	if err := name.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid name: %v", err.Error()))
	}

	var optionQuestionnaireURL types.Option[values.LauncherVersionQuestionnaireURL]
	if req.Questionnaire != nil {
		urlValue, err := url.Parse(*req.Questionnaire)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid questionnaire url")
		}

		optionQuestionnaireURL = types.NewOption(values.NewLauncherVersionQuestionnaireURL(urlValue))
	}

	gameVersionIDs := make([]values.GameVersionID, 0, len(req.GameVersions))
	for _, gameVersionID := range req.GameVersions {
		gameVersionIDs = append(gameVersionIDs, values.NewGameVersionIDFromUUID(gameVersionID))
	}

	domainEdition, err := edition.editionService.CreateEdition(
		c.Request().Context(),
		name,
		optionQuestionnaireURL,
		gameVersionIDs,
	)
	switch {
	case errors.Is(err, service.ErrInvalidGameVersionID):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid game version id")
	case errors.Is(err, service.ErrDuplicateGameVersion):
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate game version")
	case errors.Is(err, service.ErrDuplicateGame):
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate game")
	case err != nil:
		log.Printf("error: failed to create edition: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create edition")
	}

	questionnaireURL, err := domainEdition.GetQuestionnaireURL()
	if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
	}

	var strQuestionnaireURL *string
	if !errors.Is(err, domain.ErrNoQuestionnaire) {
		v := (*url.URL)(questionnaireURL).String()
		strQuestionnaireURL = &v
	}

	return c.JSON(http.StatusCreated, openapi.Edition{
		Id:            uuid.UUID(domainEdition.GetID()),
		Name:          string(domainEdition.GetName()),
		Questionnaire: strQuestionnaireURL,
		CreatedAt:     domainEdition.GetCreatedAt(),
	})
}

// エディションの削除
// (DELETE /editions/{editionID})
func (edition *Edition) DeleteEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error {
	err := edition.editionService.DeleteEdition(ctx.Request().Context(), values.NewLauncherVersionIDFromUUID(editionID))
	if errors.Is(err, service.ErrInvalidEditionID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid edition id")
	}
	if err != nil {
		log.Printf("error: failed to delete edition: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete edition")
	}

	return ctx.NoContent(http.StatusOK)
}

// エディション情報の取得
// (GET /editions/{editionID})
func (edition *Edition) GetEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error {
	domainEdition, err := edition.editionService.GetEdition(ctx.Request().Context(), values.NewLauncherVersionIDFromUUID(editionID))
	if errors.Is(err, service.ErrInvalidEditionID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid edition id")
	}
	if err != nil {
		log.Printf("error: failed to get edition: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get edition")
	}

	questionnaireURL, err := domainEdition.GetQuestionnaireURL()
	if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
	}

	var strQuestionnaireURL *string
	if !errors.Is(err, domain.ErrNoQuestionnaire) {
		v := (*url.URL)(questionnaireURL).String()
		strQuestionnaireURL = &v
	}

	return ctx.JSON(http.StatusOK, openapi.Edition{
		Id:            uuid.UUID(domainEdition.GetID()),
		Name:          string(domainEdition.GetName()),
		Questionnaire: strQuestionnaireURL,
		CreatedAt:     domainEdition.GetCreatedAt(),
	})
}

// エディション情報の変更
// (PATCH /editions/{editionID})
func (edition *Edition) PatchEdition(ctx echo.Context, editionID openapi.EditionIDInPath) error {
	var req openapi.PatchEdition
	err := ctx.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	name := values.NewLauncherVersionName(req.Name)
	if err := name.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid name: %v", err.Error()))
	}

	var optionQuestionnaireURL types.Option[values.LauncherVersionQuestionnaireURL]
	if req.Questionnaire != nil {
		urlValue, err := url.Parse(*req.Questionnaire)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid questionnaire url")
		}

		optionQuestionnaireURL = types.NewOption(values.NewLauncherVersionQuestionnaireURL(urlValue))
	}

	domainEdition, err := edition.editionService.UpdateEdition(
		ctx.Request().Context(),
		values.NewLauncherVersionIDFromUUID(editionID),
		name,
		optionQuestionnaireURL,
	)
	if errors.Is(err, service.ErrInvalidEditionID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid edition id")
	}
	if err != nil {
		log.Printf("error: failed to update edition: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update edition")
	}

	questionnaireURL, err := domainEdition.GetQuestionnaireURL()
	if err != nil && !errors.Is(err, domain.ErrNoQuestionnaire) {
		log.Printf("error: failed to get questionnaire url: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get questionnaire url")
	}

	var strQuestionnaireURL *string
	if !errors.Is(err, domain.ErrNoQuestionnaire) {
		v := (*url.URL)(questionnaireURL).String()
		strQuestionnaireURL = &v
	}

	return ctx.JSON(http.StatusOK, openapi.Edition{
		Id:            uuid.UUID(domainEdition.GetID()),
		Name:          string(domainEdition.GetName()),
		Questionnaire: strQuestionnaireURL,
		CreatedAt:     domainEdition.GetCreatedAt(),
	})
}

// エディションに紐づくゲームの一覧の取得
// (GET /editions/{editionID}/games)
func (edition *Edition) GetEditionGames(ctx echo.Context, editionID openapi.EditionIDInPath) error {
	gameVersions, err := edition.editionService.GetEditionGameVersions(ctx.Request().Context(), values.NewLauncherVersionIDFromUUID(editionID))
	if errors.Is(err, service.ErrInvalidEditionID) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid edition id")
	}
	if err != nil {
		log.Printf("error: failed to get games: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get games")
	}

	res := make([]openapi.EditionGameResponse, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		var resURL *openapi.GameURL
		urlValue, ok := gameVersion.GameVersion.Assets.URL.Value()
		if ok {
			v := (*url.URL)(urlValue).String()
			resURL = &v
		}

		var resFiles *openapi.GameVersionFiles
		windows, windowsOk := gameVersion.GameVersion.Assets.Windows.Value()
		mac, macOk := gameVersion.GameVersion.Assets.Mac.Value()
		jar, jarOk := gameVersion.GameVersion.Assets.Jar.Value()
		if windowsOk || macOk || jarOk {
			resFiles = &openapi.GameVersionFiles{}

			if windowsOk {
				v := (uuid.UUID)(windows)
				resFiles.Win32 = &v
			}

			if macOk {
				v := (uuid.UUID)(mac)
				resFiles.Darwin = &v
			}

			if jarOk {
				v := (uuid.UUID)(jar)
				resFiles.Jar = &v
			}
		}

		res = append(res, openapi.EditionGameResponse{
			Id:          uuid.UUID(gameVersion.Game.GetID()),
			Name:        string(gameVersion.Game.GetName()),
			Description: string(gameVersion.Game.GetDescription()),
			CreatedAt:   gameVersion.Game.GetCreatedAt(),
			Version: openapi.GameVersion{
				Id:          uuid.UUID(gameVersion.GameVersion.GetID()),
				Name:        string(gameVersion.GameVersion.GetName()),
				Description: string(gameVersion.GameVersion.GetDescription()),
				CreatedAt:   gameVersion.GameVersion.GetCreatedAt(),
				ImageID:     uuid.UUID(gameVersion.GameVersion.ImageID),
				VideoID:     uuid.UUID(gameVersion.GameVersion.VideoID),
				Url:         resURL,
				Files:       resFiles,
			},
		})
	}

	return ctx.JSON(http.StatusOK, res)
}

// エディションのゲームの変更
// (PATCH /editions/{editionID}/games)
func (edition *Edition) PatchEditionGame(c echo.Context, editionID openapi.EditionIDInPath) error {
	var req openapi.PatchEditionGameRequest
	err := c.Bind(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	gameVersionIDs := make([]values.GameVersionID, 0, len(req.GameVersionIDs))
	for _, gameVersionID := range req.GameVersionIDs {
		gameVersionIDs = append(gameVersionIDs, values.NewGameVersionIDFromUUID(gameVersionID))
	}

	gameVersions, err := edition.editionService.UpdateEditionGameVersions(
		c.Request().Context(),
		values.NewLauncherVersionIDFromUUID(editionID),
		gameVersionIDs,
	)
	switch {
	case errors.Is(err, service.ErrInvalidEditionID):
		return echo.NewHTTPError(http.StatusBadRequest, "invalid edition id")
	case errors.Is(err, service.ErrDuplicateGameVersion):
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate game version")
	case errors.Is(err, service.ErrDuplicateGame):
		return echo.NewHTTPError(http.StatusBadRequest, "duplicate game")
	case err != nil:
		log.Printf("error: failed to update edition games: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update edition games")
	}

	res := make([]openapi.EditionGameResponse, 0, len(gameVersions))
	for _, gameVersion := range gameVersions {
		var resURL *openapi.GameURL
		urlValue, ok := gameVersion.GameVersion.Assets.URL.Value()
		if ok {
			v := (*url.URL)(urlValue).String()
			resURL = &v
		}

		var resFiles *openapi.GameVersionFiles
		windows, windowsOk := gameVersion.GameVersion.Assets.Windows.Value()
		mac, macOk := gameVersion.GameVersion.Assets.Mac.Value()
		jar, jarOk := gameVersion.GameVersion.Assets.Jar.Value()
		if windowsOk || macOk || jarOk {
			resFiles = &openapi.GameVersionFiles{}

			if windowsOk {
				v := (uuid.UUID)(windows)
				resFiles.Win32 = &v
			}

			if macOk {
				v := (uuid.UUID)(mac)
				resFiles.Darwin = &v
			}

			if jarOk {
				v := (uuid.UUID)(jar)
				resFiles.Jar = &v
			}
		}

		res = append(res, openapi.EditionGameResponse{
			Id:          uuid.UUID(gameVersion.Game.GetID()),
			Name:        string(gameVersion.Game.GetName()),
			Description: string(gameVersion.Game.GetDescription()),
			CreatedAt:   gameVersion.Game.GetCreatedAt(),
			Version: openapi.GameVersion{
				Id:          uuid.UUID(gameVersion.GameVersion.GetID()),
				Name:        string(gameVersion.GameVersion.GetName()),
				Description: string(gameVersion.GameVersion.GetDescription()),
				CreatedAt:   gameVersion.GameVersion.GetCreatedAt(),
				ImageID:     uuid.UUID(gameVersion.GameVersion.ImageID),
				VideoID:     uuid.UUID(gameVersion.GameVersion.VideoID),
				Url:         resURL,
				Files:       resFiles,
			},
		})
	}

	return c.JSON(http.StatusOK, res)
}

// エディションのプレイ統計取得
// (GET /editions/{editionID}/play-stats)
func (edition *Edition) GetEditionPlayStats(c echo.Context, editionID openapi.EditionIDInPath, params openapi.GetEditionPlayStatsParams) error {
	// TODO: 実装が必要
	return echo.NewHTTPError(http.StatusNotImplemented, "not implemented yet")
}
