package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GameCreator struct {
	gameCreatorService service.GameCreator
}

func NewGameCreator(gameCreatorService service.GameCreator) *GameCreator {
	return &GameCreator{
		gameCreatorService: gameCreatorService,
	}
}

// ゲームクリエイターのジョブ一覧の取得
// (GET /games/{gameID}/creators/jobs)
func (gc *GameCreator) GetGameCreatorJobs(c echo.Context, gameID openapi.GameIDInPath) error {
	presentJobs, customJobs, err := gc.gameCreatorService.GetGameCreatorJobs(
		c.Request().Context(),
		values.NewGameIDFromUUID(gameID),
	)
	if errors.Is(err, service.ErrInvalidGameID) {
		return echo.NewHTTPError(http.StatusNotFound, "Invalid gameID")
	}
	if err != nil {
		log.Printf("error: failed to get game creator jobs: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get game creator jobs")
	}

	res := make([]openapi.GameCreatorJob, 0, len(presentJobs)+len(customJobs))
	for _, job := range presentJobs {
		res = append(res, openapi.GameCreatorJob{
			Id:          openapi.GameCreatorJobID(job.GetID()),
			DisplayName: openapi.GameCreatorJobDisplayName(job.GetDisplayName()),
			IsCustomJob: false,
		})
	}
	for _, job := range customJobs {
		res = append(res, openapi.GameCreatorJob{
			Id:          openapi.GameCreatorJobID(job.GetID()),
			DisplayName: openapi.GameCreatorJobDisplayName(job.GetDisplayName()),
			IsCustomJob: true,
		})
	}

	return c.JSON(http.StatusOK, res)
}

// ゲームクリエイター一覧の取得
// (GET /games/{gameID}/creators)
func (gc *GameCreator) GetGameCreators(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームクリエイターの作成
// (POST /games/{gameID}/creators)
func (gc *GameCreator) PostGameCreator(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームクリエイターの削除
// (DELETE /games/{gameID}/creators/{creatorID})
func (gc *GameCreator) DeleteGameCreator(c echo.Context, _ openapi.GameIDInPath, _ openapi.CreatorIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームクリエイターのjob更新
// (PUT /games/{gameID}/creators/{creatorID}/jobs)
func (gc *GameCreator) PutGameCreatorJobs(c echo.Context, _ openapi.GameIDInPath, _ openapi.CreatorIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}

// ゲームクリエイターのカスタムジョブ作成
// (POST /games/{gameID}/creators/custom-jobs)
func (gc *GameCreator) PostGameCreatorCustomJob(c echo.Context, _ openapi.GameIDInPath) error {
	return c.NoContent(http.StatusNotImplemented)
}
