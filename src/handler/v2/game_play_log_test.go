package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/traPtitech/trap-collection-server/src/domain"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
	"github.com/traPtitech/trap-collection-server/src/service/mock"
	"go.uber.org/mock/gomock"
)

func TestPostGamePlayLogStart(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	editionID := values.NewLauncherVersionID()
	gameID := values.NewGameID()
	gameVersionID := values.NewGameVersionID()
	gameStartTime := time.Now()
	reqBody := openapi.PostGamePlayLogStartRequest{
		GameVersionID: openapi.GameVersionID(gameVersionID),
		EditionID:     openapi.EditionID(editionID),
		GameID:        openapi.GameID(gameID),
		StartTime:     gameStartTime,
	}

	playLogID := values.NewGamePlayLogID()

	testCases := map[string]struct {
		editionID            values.LauncherVersionID
		gameID               values.GameID
		invalidReqBody       bool
		reqBody              openapi.PostGamePlayLogStartRequest
		executeCreatePlayLog bool
		playLog              *domain.GamePlayLog
		CreatePlayLogErr     error
		isError              bool
		statusCode           int
		resBody              openapi.PostGamePlayLogStartResponse
	}{
		"request bodyが不正なのでエラー": {
			editionID:      editionID,
			gameID:         gameID,
			invalidReqBody: true,
			isError:        true,
			statusCode:     http.StatusBadRequest,
		},
		"CreatePlayLogがErrInvalidEditionなので404": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     service.ErrInvalidEdition,
			isError:              true,
			statusCode:           http.StatusNotFound,
		},
		"CreatePlayLogがErrInvalidGameなので404": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     service.ErrInvalidGame,
			isError:              true,
			statusCode:           http.StatusNotFound,
		},
		"CreatePlayLogがErrInvalidGameVersionなので404": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     service.ErrInvalidGameVersion,
			isError:              true,
			statusCode:           http.StatusNotFound,
		},
		"CreatePlayLogがその他のエラーなので500": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			CreatePlayLogErr:     assert.AnError,
			isError:              true,
			statusCode:           http.StatusInternalServerError,
		},
		"CreatePlayLogが成功するので201": {
			editionID:            editionID,
			gameID:               gameID,
			reqBody:              reqBody,
			executeCreatePlayLog: true,
			playLog:              domain.NewGamePlayLog(playLogID, editionID, gameID, gameVersionID, gameStartTime, nil, time.Now(), time.Now()),
			statusCode:           http.StatusCreated,
			resBody: openapi.PostGamePlayLogStartResponse{
				PlayLogID: openapi.GamePlayLogID(playLogID),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			serviceMock := mock.NewMockGamePlayLogV2(ctrl)
			h := NewGamePlayLog(serviceMock)

			gameVersionID := values.NewGameVersionIDFromUUID(testCase.reqBody.GameVersionID)

			if testCase.executeCreatePlayLog {
				serviceMock.
					EXPECT().
					CreatePlayLog(
						gomock.Any(),
						testCase.editionID,
						testCase.gameID,
						gameVersionID,
						gomock.Cond(func(startTime time.Time) bool { return startTime.Sub(testCase.reqBody.StartTime).Abs() < time.Second }), // JSONのエンコードとデコードで精度がずれるため
					).
					Return(testCase.playLog, testCase.CreatePlayLogErr)
			}

			var body bodyOpt
			if testCase.invalidReqBody {
				body = withStringBody(t, "invalid")
			} else {
				body = withJSONBody(t, testCase.reqBody)
			}

			url := fmt.Sprintf("/editions/%s/games/%s/plays/start",
				uuid.UUID(testCase.editionID).String(), uuid.UUID(testCase.gameID).String())
			c, _, rec := setupTestRequest(t, http.MethodPost, url, body)

			err := h.PostGamePlayLogStart(c, openapi.EditionIDInPath(testCase.editionID), openapi.GameIDInPath(testCase.gameID))

			if testCase.isError {
				var httpError *echo.HTTPError
				assert.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.statusCode, httpError.Code)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			var resBody openapi.PostGamePlayLogStartResponse
			assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resBody))
			assert.Equal(t, testCase.resBody.PlayLogID, resBody.PlayLogID)
		})
	}

}

func TestPatchGamePlayLogEnd(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	editionID := values.NewLauncherVersionID()
	gameID := values.NewGameID()
	playLogID := values.NewGamePlayLogID()
	endTime := time.Now()
	reqBody := openapi.PatchGamePlayLogEndRequest{
		EndTime: endTime,
	}

	testCases := map[string]struct {
		editionID                   values.LauncherVersionID
		gameID                      values.GameID
		playLogID                   values.GamePlayLogID
		invalidReqBody              bool
		reqBody                     openapi.PatchGamePlayLogEndRequest
		executeUpdatePlayLogEndTime bool
		UpdatePlayLogEndTimeErr     error
		isError                     bool
		statusCode                  int
	}{
		"request bodyが不正なのでエラー": {
			editionID:      editionID,
			gameID:         gameID,
			playLogID:      playLogID,
			invalidReqBody: true,
			isError:        true,
			statusCode:     http.StatusBadRequest,
		},
		"UpdatePlayLogEndTimeがErrInvalidPlayLogIDなので404": {
			editionID:                   editionID,
			gameID:                      gameID,
			playLogID:                   playLogID,
			reqBody:                     reqBody,
			executeUpdatePlayLogEndTime: true,
			UpdatePlayLogEndTimeErr:     service.ErrInvalidPlayLogID,
			isError:                     true,
			statusCode:                  http.StatusNotFound,
		},
		"UpdatePlayLogEndTimeがErrInvalidEndTimeなので400": {
			editionID:                   editionID,
			gameID:                      gameID,
			playLogID:                   playLogID,
			reqBody:                     reqBody,
			executeUpdatePlayLogEndTime: true,
			UpdatePlayLogEndTimeErr:     service.ErrInvalidEndTime,
			isError:                     true,
			statusCode:                  http.StatusBadRequest,
		},
		"UpdatePlayLogEndTimeがその他のエラーなので500": {
			editionID:                   editionID,
			gameID:                      gameID,
			playLogID:                   playLogID,
			reqBody:                     reqBody,
			executeUpdatePlayLogEndTime: true,
			UpdatePlayLogEndTimeErr:     assert.AnError,
			isError:                     true,
			statusCode:                  http.StatusInternalServerError,
		},
		"UpdatePlayLogEndTimeがErrInvalidPlayLogEditionGamePairなので400": {
			editionID:                   editionID,
			gameID:                      gameID,
			playLogID:                   playLogID,
			reqBody:                     reqBody,
			executeUpdatePlayLogEndTime: true,
			UpdatePlayLogEndTimeErr:     service.ErrInvalidPlayLogEditionGamePair,
			isError:                     true,
			statusCode:                  http.StatusBadRequest,
		},
		"UpdatePlayLogEndTimeが成功するので200": {
			editionID:                   editionID,
			gameID:                      gameID,
			playLogID:                   playLogID,
			reqBody:                     reqBody,
			executeUpdatePlayLogEndTime: true,
			statusCode:                  http.StatusOK,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			serviceMock := mock.NewMockGamePlayLogV2(ctrl)
			h := NewGamePlayLog(serviceMock)

			if testCase.executeUpdatePlayLogEndTime {
				serviceMock.
					EXPECT().
					UpdatePlayLogEndTime(
						gomock.Any(),
						testCase.editionID,
						testCase.gameID,
						testCase.playLogID,
						gomock.Cond(func(endTime time.Time) bool { return endTime.Sub(testCase.reqBody.EndTime).Abs() < time.Second }),
					).
					Return(testCase.UpdatePlayLogEndTimeErr)
			}

			var body bodyOpt
			if testCase.invalidReqBody {
				body = withStringBody(t, "invalid")
			} else {
				body = withJSONBody(t, testCase.reqBody)
			}

			url := fmt.Sprintf("/editions/%s/games/%s/plays/%s/end",
				uuid.UUID(testCase.editionID).String(), uuid.UUID(testCase.gameID).String(), uuid.UUID(testCase.playLogID).String())
			c, _, rec := setupTestRequest(t, http.MethodPatch, url, body)

			err := h.PatchGamePlayLogEnd(c, openapi.EditionIDInPath(testCase.editionID), openapi.GameIDInPath(testCase.gameID), openapi.PlayLogIDInPath(testCase.playLogID))

			if testCase.isError {
				var httpError *echo.HTTPError
				assert.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.statusCode, httpError.Code)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)
		})
	}
}

func TestGetEditionPlayStats(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	editionID := values.NewLauncherVersionID()
	editionName := "Test Edition"
	gameID1 := values.NewGameID()
	gameID2 := values.NewGameID()

	now := time.Now()
	defaultStart := now.Add(-24 * time.Hour)
	customStart := now.Add(-48 * time.Hour)
	customEnd := now.Add(-24 * time.Hour)

	gameStats := []*domain.GamePlayStatsInEdition{
		domain.NewGamePlayStatsInEdition(gameID1, 10, 3600*time.Second),
		domain.NewGamePlayStatsInEdition(gameID2, 5, 1800*time.Second),
	}
	hourlyStats := []*domain.HourlyPlayStats{
		domain.NewHourlyPlayStats(customStart, 3, 900*time.Second),
		domain.NewHourlyPlayStats(customStart.Add(time.Hour), 5, 1500*time.Second),
		domain.NewHourlyPlayStats(customStart.Add(2*time.Hour), 7, 2000*time.Second),
	}

	editionStats := domain.NewEditionPlayStats(
		editionID,
		values.NewLauncherVersionName(editionName),
		15,
		5400*time.Second,
		gameStats,
		hourlyStats,
	)

	expectedGameStats := []openapi.GamePlayStatsInEdition{
		{
			GameID:    openapi.GameID(gameID1),
			PlayCount: 10,
			PlayTime:  3600,
		},
		{
			GameID:    openapi.GameID(gameID2),
			PlayCount: 5,
			PlayTime:  1800,
		},
	}

	expectedHourlyStats := []openapi.HourlyPlayStats{
		{
			StartTime: customStart,
			PlayCount: 3,
			PlayTime:  900,
		},
		{
			StartTime: customStart.Add(time.Hour),
			PlayCount: 5,
			PlayTime:  1500,
		},
		{
			StartTime: customStart.Add(2 * time.Hour),
			PlayCount: 7,
			PlayTime:  2000,
		},
	}

	expectedEditionPlayStats := openapi.EditionPlayStats{
		EditionID:        openapi.EditionID(editionID),
		EditionName:      editionName,
		TotalPlayCount:   15,
		TotalPlaySeconds: 5400,
		GameStats:       expectedGameStats,
		HourlyStats:     expectedHourlyStats,
	}

	testCases := map[string]struct {
		editionID               values.LauncherVersionID
		queryParams            map[string]string
		executeGetEditionStats bool
		expectedStart          time.Time
		expectedEnd            time.Time
		editionStats          *domain.EditionPlayStats
		getEditionStatsErr    error
		expectedResponse      openapi.EditionPlayStats
		isError               bool
		statusCode            int
	}{
		"クエリパラメータなしでエラーなし": {
			editionID:              editionID,
			queryParams:           map[string]string{},
			executeGetEditionStats: true,
			expectedStart:          defaultStart,
			expectedEnd:            now,
			editionStats:          editionStats,
			expectedResponse: expectedEditionPlayStats,
			statusCode: http.StatusOK,
		},
		"start/endの両方を指定してもエラーなし": {
			editionID: editionID,
			queryParams: map[string]string{
				"start": customStart.Format(time.RFC3339),
				"end":   customEnd.Format(time.RFC3339),
			},
			executeGetEditionStats: true,
			expectedStart:          customStart,
			expectedEnd:            customEnd,
			editionStats:          editionStats,
			expectedResponse: expectedEditionPlayStats,
			statusCode: http.StatusOK,
		},
		"startのみ指定でもエラーなし": {
			editionID: editionID,
			queryParams: map[string]string{
				"start": customStart.Format(time.RFC3339),
			},
			executeGetEditionStats: true,
			expectedStart:          customStart,
			expectedEnd:            now,
			editionStats:          editionStats,
			expectedResponse: expectedEditionPlayStats,
			statusCode: http.StatusOK,
		},
		"endのみ指定でもエラーなし": {
			editionID: editionID,
			queryParams: map[string]string{
				"end": customEnd.Format(time.RFC3339),
			},
			executeGetEditionStats: true,
			expectedStart:          customEnd.Add(-24 * time.Hour),
			expectedEnd:            customEnd,
			editionStats:          editionStats,
			expectedResponse: expectedEditionPlayStats,
			statusCode: http.StatusOK,
		},
		"GetEditionPlayStatsがErrInvalidEditionなので404": {
			editionID:              editionID,
			queryParams:           map[string]string{},
			executeGetEditionStats: true,
			expectedStart:          defaultStart,
			expectedEnd:            now,
			getEditionStatsErr:    service.ErrInvalidEdition,
			isError:               true,
			statusCode:            http.StatusNotFound,
		},
		"GetEditionPlayStatsがその他のエラーなので500": {
			editionID:              editionID,
			queryParams:           map[string]string{},
			executeGetEditionStats: true,
			expectedStart:          defaultStart,
			expectedEnd:            now,
			getEditionStatsErr:    assert.AnError,
			isError:               true,
			statusCode:            http.StatusInternalServerError,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			serviceMock := mock.NewMockGamePlayLogV2(ctrl)
			h := NewGamePlayLog(serviceMock)

			if testCase.executeGetEditionStats {
				serviceMock.
					EXPECT().
					GetEditionPlayStats(
						gomock.Any(),
						testCase.editionID,
						gomock.Cond(func(start time.Time) bool {
							return start.Sub(testCase.expectedStart).Abs() < time.Hour
						}),
						gomock.Cond(func(end time.Time) bool {
							return end.Sub(testCase.expectedEnd).Abs() < time.Hour
						}),
					).
					Return(testCase.editionStats, testCase.getEditionStatsErr)
			}

			url := fmt.Sprintf("/editions/%s/play-stats", uuid.UUID(testCase.editionID).String())
			if len(testCase.queryParams) > 0 {
				url += "?"
				first := true
				for key, value := range testCase.queryParams {
					if !first {
						url += "&"
					}
					url += fmt.Sprintf("%s=%s", key, value)
					first = false
				}
			}

			c, _, rec := setupTestRequest(t, http.MethodGet, url, nil)

			params := openapi.GetEditionPlayStatsParams{}
			if start, ok := testCase.queryParams["start"]; ok {
				startTime, _ := time.Parse(time.RFC3339, start)
				params.Start = &startTime
			}
			if end, ok := testCase.queryParams["end"]; ok {
				endTime, _ := time.Parse(time.RFC3339, end)
				params.End = &endTime
			}

			err := h.GetEditionPlayStats(c, openapi.EditionIDInPath(testCase.editionID), params)

			if testCase.isError {
				var httpError *echo.HTTPError
				assert.ErrorAs(t, err, &httpError)
				assert.Equal(t, testCase.statusCode, httpError.Code)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.statusCode, rec.Code)

			if !testCase.isError {
				var resBody openapi.EditionPlayStats
				err = json.NewDecoder(rec.Body).Decode(&resBody)
				assert.NoError(t, err)

				assert.Equal(t, testCase.expectedResponse.EditionID, resBody.EditionID)
				assert.Equal(t, testCase.expectedResponse.EditionName, resBody.EditionName)
				assert.Equal(t, testCase.expectedResponse.TotalPlayCount, resBody.TotalPlayCount)
				assert.Equal(t, testCase.expectedResponse.TotalPlaySeconds, resBody.TotalPlaySeconds)

				assert.Len(t, resBody.GameStats, len(testCase.expectedResponse.GameStats))
				for i, expectedGame := range testCase.expectedResponse.GameStats {
					assert.Equal(t, expectedGame.GameID, resBody.GameStats[i].GameID)
					assert.Equal(t, expectedGame.PlayCount, resBody.GameStats[i].PlayCount)
					assert.Equal(t, expectedGame.PlayTime, resBody.GameStats[i].PlayTime)
				}

				assert.Len(t, resBody.HourlyStats, len(testCase.expectedResponse.HourlyStats))
				for i, expectedHourly := range testCase.expectedResponse.HourlyStats {
					assert.WithinDuration(t, expectedHourly.StartTime, resBody.HourlyStats[i].StartTime, time.Second)
					assert.Equal(t, expectedHourly.PlayCount, resBody.HourlyStats[i].PlayCount)
					assert.Equal(t, expectedHourly.PlayTime, resBody.HourlyStats[i].PlayTime)
				}
			}
		})
	}
}
