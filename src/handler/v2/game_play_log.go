package v2

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type GamePlayLog struct {
	gamePlayLogService service.GamePlayLogV2
}

func NewGamePlayLog(gamePlayLogService service.GamePlayLogV2) *GamePlayLog {
	return &GamePlayLog{
		gamePlayLogService: gamePlayLogService,
	}
}

// ゲーム起動ログの記録
// (POST /editions/{editionID}/games/{gameID}/plays/start)
func (gpl *GamePlayLog) PostGamePlayLogStart(c echo.Context, editionIDPath openapi.EditionIDInPath, gameIDPath openapi.GameIDInPath) error {
	ctx := c.Request().Context()
	var body openapi.PostGamePlayLogStartJSONRequestBody
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	editionID := values.NewEditionIDFromUUID(editionIDPath)
	gameID := values.NewGameIDFromUUID(gameIDPath)
	gameVersionID := values.NewGameVersionIDFromUUID(body.GameVersionID)
	startAt := body.StartTime

	playLog, err := gpl.gamePlayLogService.CreatePlayLog(ctx, editionID, gameID, gameVersionID, startAt)
	if errors.Is(err, service.ErrInvalidEdition) {
		return echo.NewHTTPError(http.StatusNotFound, "edition not found")
	}
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if errors.Is(err, service.ErrInvalidGameVersion) {
		return echo.NewHTTPError(http.StatusNotFound, "game version not found")
	}
	if err != nil {
		log.Printf("error: failed to create game play log: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to post game play log start")
	}

	res := openapi.PostGamePlayLogStartResponse{
		PlayLogID: openapi.GamePlayLogID(playLog.GetID()),
	}
	return c.JSON(http.StatusCreated, res)
}

// ゲーム終了ログの記録
// (PATCH /editions/{editionID}/games/{gameID}/plays/{playLogID}/end)
func (gpl *GamePlayLog) PatchGamePlayLogEnd(c echo.Context, editionIDPath openapi.EditionIDInPath, gameIDPath openapi.GameIDInPath, playLogIDPath openapi.PlayLogIDInPath) error {
	ctx := c.Request().Context()
	var body openapi.PatchGamePlayLogEndJSONRequestBody
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}

	editionID := values.NewEditionIDFromUUID(editionIDPath)
	gameID := values.NewGameIDFromUUID(gameIDPath)
	playLogID := values.GamePlayLogIDFromUUID(uuid.UUID(playLogIDPath))
	endTime := body.EndTime

	err := gpl.gamePlayLogService.UpdatePlayLogEndTime(ctx, editionID, gameID, playLogID, endTime)
	if errors.Is(err, service.ErrInvalidPlayLogID) {
		return echo.NewHTTPError(http.StatusNotFound, "play log not found")
	}
	if errors.Is(err, service.ErrInvalidEndTime) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid end time")
	}
	if errors.Is(err, service.ErrInvalidPlayLogEditionGamePair) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid play log edition and game pair")
	}
	if err != nil {
		log.Printf("error: failed to update game play log end time: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to patch game play log end")
	}

	return c.NoContent(http.StatusOK)
}

// ゲームプレイ統計の取得
// (GET /games/{gameID}/play-stats)
func (gpl *GamePlayLog) GetGamePlayStats(c echo.Context, gameIDPath openapi.GameIDInPath, params openapi.GetGamePlayStatsParams) error {
	ctx := c.Request().Context()

	gameID := values.NewGameIDFromUUID(gameIDPath)

	var gameVersionID *values.GameVersionID
	if params.GameVersionID != nil {
		vID := values.NewGameVersionIDFromUUID(*params.GameVersionID)
		gameVersionID = &vID
	}

	var start, end time.Time
	if params.End != nil {
		end = *params.End
	} else {
		end = time.Now()
	}
	if params.Start != nil {
		start = *params.Start
	} else {
		start = end.Add(-24 * time.Hour)
	}

	// Serviceの呼び出し
	stats, err := gpl.gamePlayLogService.GetGamePlayStats(ctx, gameID, gameVersionID, start, end)
	if errors.Is(err, service.ErrInvalidGame) {
		return echo.NewHTTPError(http.StatusNotFound, "game not found")
	}
	if errors.Is(err, service.ErrInvalidTimeRange) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid time range")
	}
	if errors.Is(err, service.ErrTimePeriodTooLong) {
		return echo.NewHTTPError(http.StatusBadRequest, "time period too long")
	}
	if err != nil {
		log.Printf("get game play stats: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "get game play stats")
	}

	hourlyStats := make([]openapi.HourlyPlayStats, 0, len(stats.GetHourlyStats()))
	for _, hourlyStat := range stats.GetHourlyStats() {
		hourlyStats = append(hourlyStats, openapi.HourlyPlayStats{
			StartTime: hourlyStat.GetStartTime(),
			PlayCount: hourlyStat.GetPlayCount(),
			PlayTime:  int(hourlyStat.GetPlayTime().Seconds()),
		})
	}

	res := openapi.GamePlayStats{
		GameID:           uuid.UUID(stats.GetGameID()),
		TotalPlayCount:   stats.GetTotalPlayCount(),
		TotalPlaySeconds: int(stats.GetTotalPlayTime().Seconds()),
		HourlyStats:      hourlyStats,
	}

	return c.JSON(http.StatusOK, res)
}

// エディションプレイ統計の取得
// (GET /editions/{editionID}/play-stats)
func (gpl *GamePlayLog) GetEditionPlayStats(c echo.Context, editionIDPath openapi.EditionIDInPath, params openapi.GetEditionPlayStatsParams) error {
	ctx := c.Request().Context()

	editionID := values.NewEditionIDFromUUID(editionIDPath)

	var start, end time.Time

	if params.End != nil {
		end = *params.End
	} else {
		end = time.Now()
	}
	if params.Start != nil {
		start = *params.Start
	} else {
		start = end.Add(-24 * time.Hour)
	}

	stats, err := gpl.gamePlayLogService.GetEditionPlayStats(ctx, editionID, start, end)
	if errors.Is(err, service.ErrInvalidEdition) {
		return echo.NewHTTPError(http.StatusNotFound, "edition not found")
	}
	if errors.Is(err, service.ErrInvalidTimeRange) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid time range")
	}
	if errors.Is(err, service.ErrTimePeriodTooLong) {
		return echo.NewHTTPError(http.StatusBadRequest, "time period too long")
	}
	if err != nil {
		log.Printf("error: failed to get edition play stats: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get edition play stats")
	}

	res := openapi.EditionPlayStats{
		EditionID:        openapi.EditionID(stats.GetEditionID()),
		EditionName:      string(stats.GetEditionName()),
		TotalPlayCount:   stats.GetTotalPlayCount(),
		TotalPlaySeconds: int(stats.GetTotalPlayTime().Seconds()),
		GameStats:        make([]openapi.GamePlayStatsInEdition, 0, len(stats.GetGameStats())),
		HourlyStats:      make([]openapi.HourlyPlayStats, 0, len(stats.GetHourlyStats())),
	}

	for _, gameStat := range stats.GetGameStats() {
		res.GameStats = append(res.GameStats, openapi.GamePlayStatsInEdition{
			GameID:    openapi.GameID(gameStat.GetGameID()),
			PlayCount: gameStat.GetPlayCount(),
			PlayTime:  int(gameStat.GetPlayTime().Seconds()),
		})
	}
	for _, hourlyStat := range stats.GetHourlyStats() {
		res.HourlyStats = append(res.HourlyStats, openapi.HourlyPlayStats{
			StartTime: hourlyStat.GetStartTime(),
			PlayCount: hourlyStat.GetPlayCount(),
			PlayTime:  int(hourlyStat.GetPlayTime().Seconds()),
		})
	}

	return c.JSON(http.StatusOK, res)
}
